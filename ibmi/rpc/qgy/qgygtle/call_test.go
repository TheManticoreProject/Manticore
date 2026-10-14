package qgygtle_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgygtle"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

var codec, _ = ccsid.Lookup(ccsid.CCSID37)

type fakeCaller struct {
	respond func(rpc.Call) (*rpc.Result, error)
}

func (f fakeCaller) Call(ctx context.Context, c rpc.Call) (*rpc.Result, error) { return f.respond(c) }

var handle = openlist.Handle{1, 2, 3, 4}

func info(total, returned int32, h openlist.Handle, recLen, first, infoLen int32) []byte {
	b := make([]byte, openlist.Length)
	binary.BigEndian.PutUint32(b[0:], uint32(total))
	binary.BigEndian.PutUint32(b[4:], uint32(returned))
	copy(b[8:], h[:])
	binary.BigEndian.PutUint32(b[12:], uint32(recLen))
	b[16] = 0xC3
	b[30] = 0xF2
	binary.BigEndian.PutUint32(b[32:], uint32(infoLen))
	binary.BigEndian.PutUint32(b[36:], uint32(first))
	return b
}

func respond(receiver, li []byte) func(rpc.Call) (*rpc.Result, error) {
	return func(c rpc.Call) (*rpc.Result, error) {
		r := &rpc.Result{Parameters: make([]rpc.ParameterResult, c.ParameterCount()), RequestedMessageOption: c.MessageOption(), EffectiveMessageOption: c.MessageOption()}
		r.Parameters[0] = rpc.ParameterResult{Present: true, Data: receiver}
		r.Parameters[3] = rpc.ParameterResult{Present: true, Data: li}
		return r, nil
	}
}

var req = qgygtle.Request{Handle: handle, RecordLength: 62, Count: 2, Start: 3}

func TestNewCallParameters(t *testing.T) {
	call, err := qgygtle.NewCall(req)
	if err != nil {
		t.Fatal(err)
	}
	if call.Library().String() != "QGY" || call.Program().String() != "QGYGTLE" {
		t.Errorf("call = %v", call)
	}
	want := []struct {
		dir   rpc.Direction
		max   uint32
		input []byte
	}{
		{rpc.DirectionOutput, 124, nil},
		{rpc.DirectionInput, 4, []byte{0, 0, 0, 124}},
		{rpc.DirectionInput, 4, handle[:]},
		{rpc.DirectionOutput, 80, nil},
		{rpc.DirectionInput, 4, []byte{0, 0, 0, 2}},
		{rpc.DirectionInput, 4, []byte{0, 0, 0, 3}},
		{rpc.DirectionInput, 8, make([]byte, 8)},
	}
	for i, w := range want {
		p := call.Parameters()[i]
		if p.Direction() != w.dir || p.Maximum() != w.max || !bytes.Equal(p.InputBytes(), w.input) {
			t.Errorf("parameter %d = %v % x", i+1, p, p.InputBytes())
		}
	}
}

func TestNewCallRejects(t *testing.T) {
	for name, r := range map[string]qgygtle.Request{
		"zero count":        {RecordLength: 62, Count: 0, Start: 1},
		"zero start":        {RecordLength: 62, Count: 1, Start: 0},
		"zero length":       {RecordLength: 0, Count: 1, Start: 1},
		"receiver overflow": {RecordLength: 228, Count: 0x7FFFFFFF, Start: 1},
	} {
		if _, err := qgygtle.NewCall(r); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestGet(t *testing.T) {
	receiver := bytes.Repeat([]byte{0xAB}, 124)
	page, err := qgygtle.Get(context.Background(), fakeCaller{respond(receiver, info(63, 2, handle, 62, 3, 124))}, req, codec)
	if err != nil || len(page.Records) != 124 || page.Info.RecordsReturned != 2 {
		t.Fatalf("Get = %+v, %v", page, err)
	}
	// A short final page returns fewer records than requested.
	page, err = qgygtle.Get(context.Background(), fakeCaller{respond(receiver, info(4, 1, handle, 62, 3, 62))}, req, codec)
	if err != nil || len(page.Records) != 62 {
		t.Fatalf("short page = %+v, %v", page, err)
	}
	// No record: the first-record field is not checked.
	if _, err := qgygtle.Get(context.Background(), fakeCaller{respond(receiver, info(2, 0, handle, 62, 0, 0))}, req, codec); err != nil {
		t.Errorf("empty page = %v", err)
	}
}

func TestGetRejectsInconsistentPages(t *testing.T) {
	receiver := make([]byte, 124)
	for name, li := range map[string][]byte{
		"handle changed":   info(63, 2, openlist.Handle{9, 9, 9, 9}, 62, 3, 124),
		"record length":    info(63, 2, handle, 176, 3, 124),
		"too many records": info(63, 3, handle, 62, 3, 186),
		"wrong first":      info(63, 2, handle, 62, 4, 124),
		"short info":       info(63, 2, handle, 62, 3, 100),
	} {
		if _, err := qgygtle.Get(context.Background(), fakeCaller{respond(receiver, li)}, req, codec); !errors.Is(err, qgygtle.ErrInvalidPage) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := qgygtle.Get(context.Background(), fakeCaller{respond(make([]byte, 100), info(63, 2, handle, 62, 3, 124))}, req, codec); !errors.Is(err, qgygtle.ErrInvalidPage) {
		t.Errorf("short receiver: %v", err)
	}
}
