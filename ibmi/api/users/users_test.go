package users_test

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/api/internal/openlist"
	"github.com/TheManticoreProject/Manticore/ibmi/api/users"
	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	listinfo "github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

var codec, _ = ccsid.Lookup(ccsid.CCSID37)

func ebcdic(s string, n int) []byte {
	b, err := ccsid.EncodeFixed(codec, s, n)
	if err != nil {
		panic(err)
	}
	return b
}

type profile struct {
	name, desc     string
	group, members bool
}

func (p profile) record() []byte {
	ind := func(b bool) byte {
		if b {
			return 0xF1
		}
		return 0xF0
	}
	return append(append(ebcdic(p.name, 10), ind(p.group), ind(p.members)), ebcdic(p.desc, 50)...)
}

// fakeIBMi serves an AUTU0150 list behind rpc.ExclusiveCaller.
type fakeIBMi struct {
	profiles []profile
	mu       sync.Mutex
	calls    []rpc.Call
}

func (f *fakeIBMi) ExecuteExclusive(ctx context.Context, fn func(rpc.Caller) error) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return fn(f)
}

func listInfo(total, returned, first int) []byte {
	b := make([]byte, listinfo.Length)
	binary.BigEndian.PutUint32(b[0:], uint32(total))
	binary.BigEndian.PutUint32(b[4:], uint32(returned))
	copy(b[8:], []byte{7, 7, 7, 7})
	binary.BigEndian.PutUint32(b[12:], 62)
	b[16] = 0xC3
	b[30] = 0xF2
	binary.BigEndian.PutUint32(b[32:], uint32(returned*62))
	binary.BigEndian.PutUint32(b[36:], uint32(first))
	return b
}

func (f *fakeIBMi) Call(ctx context.Context, call rpc.Call) (*rpc.Result, error) {
	f.calls = append(f.calls, call)
	res := &rpc.Result{Parameters: make([]rpc.ParameterResult, call.ParameterCount()), RequestedMessageOption: call.MessageOption(), EffectiveMessageOption: call.MessageOption()}
	p := call.Parameters()
	switch call.Program().String() {
	case "QGYOLAUS":
		res.Parameters[2] = rpc.ParameterResult{Present: true, Data: listInfo(len(f.profiles), 0, 0)}
	case "QGYGTLE":
		count := int(binary.BigEndian.Uint32(p[4].InputBytes()))
		start := int(binary.BigEndian.Uint32(p[5].InputBytes()))
		n := min(count, len(f.profiles)-start+1)
		receiver := make([]byte, p[0].Maximum())
		for i := 0; i < n; i++ {
			copy(receiver[i*62:], f.profiles[start-1+i].record())
		}
		res.Parameters[0] = rpc.ParameterResult{Present: true, Data: receiver}
		res.Parameters[3] = rpc.ParameterResult{Present: true, Data: listInfo(len(f.profiles), n, start)}
	}
	return res, nil
}

func newClient(t *testing.T, f *fakeIBMi) *users.Client {
	t.Helper()
	c, err := users.New(f, users.Options{SystemCCSID: 37, JobCCSID: 37})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestList(t *testing.T) {
	f := &fakeIBMi{}
	for i := 0; i < 7; i++ {
		f.profiles = append(f.profiles, profile{name: fmt.Sprintf("USER%d", i), desc: fmt.Sprintf("User number %d", i)})
	}
	f.profiles[3] = profile{name: "DEVGRP", desc: "Developers", group: true, members: true}
	res, err := newClient(t, f).List(context.Background(), users.ListOptions{PageSize: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Users) != 7 || res.TotalReported != 7 || res.Partial || res.SnapshotStatus != listinfo.StatusComplete || res.Selection != users.AllProfiles {
		t.Fatalf("result = %+v", res)
	}
	if res.Users[0] != (users.UserSummary{Name: "USER0", Description: "User number 0"}) {
		t.Errorf("user 0 = %+v", res.Users[0])
	}
	if g := res.Users[3]; g.Name != "DEVGRP" || !g.GroupProfile || !g.GroupHasMembers || g.Description != "Developers" {
		t.Errorf("group = %+v", g)
	}
	if len(res.Warnings) != 1 || res.Warnings[0] != users.AuthorityCaveat {
		t.Errorf("warnings = %q; the authority caveat is always reported", res.Warnings)
	}
	var progs []string
	for _, c := range f.calls {
		progs = append(progs, c.Program().String())
	}
	if fmt.Sprint(progs) != "[QGYOLAUS QGYGTLE QGYGTLE QGYGTLE QGYCLST]" {
		t.Errorf("calls = %v", progs)
	}
	if f := f.calls[0].Parameters()[4].InputBytes(); string(mustDecode(f)) != "AUTU0150" {
		t.Errorf("format = %q; List defaults to AUTU0150", mustDecode(f))
	}
}

func mustDecode(b []byte) string {
	s, _ := codec.Decode(b)
	return s
}

func TestListSelectionEncoding(t *testing.T) {
	for _, tc := range []struct {
		opts          users.ListOptions
		sel, grp, prf string
	}{
		{users.ListOptions{}, "*ALL", "*NONE", "*ALL"},
		{users.ListOptions{Selection: users.UserProfiles}, "*USER", "*NONE", "*ALL"},
		{users.ListOptions{Selection: users.GroupProfiles, ProfilePattern: "DEV*"}, "*GROUP", "*NONE", "DEV*"},
		{users.ListOptions{Selection: users.GroupMembers, Group: "devgrp"}, "*MEMBER", "DEVGRP", "*ALL"},
	} {
		f := &fakeIBMi{}
		if _, err := newClient(t, f).List(context.Background(), tc.opts); err != nil {
			t.Fatal(err)
		}
		p := f.calls[0].Parameters()
		got := [3]string{mustDecode(p[5].InputBytes()), mustDecode(p[6].InputBytes()), mustDecode(p[8].InputBytes())}
		want := [3]string{tc.sel, tc.grp, tc.prf}
		for i := range got {
			if strings.TrimRight(got[i], " ") != want[i] {
				t.Errorf("%+v: parameters = %q, want %q", tc.opts, got, want)
			}
		}
	}
}

func TestListRejectsBadOptionsBeforeCalling(t *testing.T) {
	for name, opts := range map[string]users.ListOptions{
		"members without group": {Selection: users.GroupMembers},
		"group without members": {Group: "DEV"},
		"bad pattern":           {ProfilePattern: "A B"},
		"undefined selection":   {Selection: 9},
		"page size":             {PageSize: openlist.MaxPageSize + 1},
	} {
		f := &fakeIBMi{}
		if _, err := newClient(t, f).List(context.Background(), opts); err == nil || len(f.calls) != 0 {
			t.Errorf("%s: error %v after %d calls", name, err, len(f.calls))
		}
	}
}

func TestListPartialOnLimit(t *testing.T) {
	f := &fakeIBMi{}
	for i := 0; i < 10; i++ {
		f.profiles = append(f.profiles, profile{name: fmt.Sprintf("U%d", i)})
	}
	res, err := newClient(t, f).List(context.Background(), users.ListOptions{PageSize: 4, MaxUsers: 6})
	if !errors.Is(err, openlist.ErrLimit) || res == nil || !res.Partial || len(res.Users) != 6 || res.TotalReported != 10 {
		t.Fatalf("result %+v, error %v", res, err)
	}
	if res.Users[5].Name != "U5" {
		t.Errorf("partial users = %+v", res.Users)
	}
}

func TestNewValidatesCCSIDs(t *testing.T) {
	f := &fakeIBMi{}
	for name, opts := range map[string]users.Options{
		"no system CCSID": {JobCCSID: 37},
		"no job CCSID":    {SystemCCSID: 37},
		"unsupported":     {SystemCCSID: 273, JobCCSID: 37},
	} {
		if _, err := users.New(f, opts); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if _, err := users.New(nil, users.Options{SystemCCSID: 37, JobCCSID: 37}); err == nil {
		t.Error("nil caller accepted")
	}
}
