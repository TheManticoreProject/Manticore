package qgyolaus_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyolaus"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyolaus/structures"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/errorcode"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

var codec, _ = ccsid.Lookup(ccsid.CCSID37)

func ebcdic(s string, n int) []byte {
	b, _ := ccsid.EncodeFixed(codec, s, n)
	return b
}

// fakeCaller answers every call with respond and records the calls.
type fakeCaller struct {
	calls   []rpc.Call
	respond func(rpc.Call) (*rpc.Result, error)
}

func (f *fakeCaller) Call(ctx context.Context, c rpc.Call) (*rpc.Result, error) {
	f.calls = append(f.calls, c)
	return f.respond(c)
}

// listInfo builds 80 bytes of list information.
func listInfo(total, returned int32, handle [4]byte, recLen int32, status byte) []byte {
	b := make([]byte, openlist.Length)
	binary.BigEndian.PutUint32(b[0:], uint32(total))
	binary.BigEndian.PutUint32(b[4:], uint32(returned))
	copy(b[8:], handle[:])
	binary.BigEndian.PutUint32(b[12:], uint32(recLen))
	b[16] = 0xC3
	b[30] = status
	return b
}

// result fills output-capable slots: list information at 2, error code at
// 7 when it is input/output.
func result(c rpc.Call, info []byte, errc []byte) *rpc.Result {
	r := &rpc.Result{Parameters: make([]rpc.ParameterResult, c.ParameterCount()), RequestedMessageOption: c.MessageOption(), EffectiveMessageOption: c.MessageOption()}
	r.Parameters[2] = rpc.ParameterResult{Present: true, Data: info}
	if c.Parameters()[7].Direction() == rpc.DirectionInputOutput {
		r.Parameters[7] = rpc.ParameterResult{Present: true, Data: errc}
	}
	return r
}

var handle = [4]byte{1, 2, 3, 4}

func TestNewCallParameters(t *testing.T) {
	call, err := qgyolaus.NewCall(qgyolaus.AUTU0150(), qgyolaus.Request{}, codec)
	if err != nil {
		t.Fatal(err)
	}
	if call.Library().String() != "QGY" || call.Program().String() != "QGYOLAUS" || call.MessageOption() != rpc.MessagesUpToTen {
		t.Errorf("call = %v option %s", call, call.MessageOption())
	}
	want := []struct {
		dir   rpc.Direction
		max   uint32
		input []byte
	}{
		{rpc.DirectionInput, 0, []byte{}}, // zero-length receiver (JTOpen 4.8 wire form)
		{rpc.DirectionInput, 4, []byte{0, 0, 0, 0}},
		{rpc.DirectionOutput, 80, nil},
		{rpc.DirectionInput, 4, []byte{0xFF, 0xFF, 0xFF, 0xFF}}, // -1: synchronous
		{rpc.DirectionInput, 8, ebcdic("AUTU0150", 8)},
		{rpc.DirectionInput, 10, ebcdic("*ALL", 10)},
		{rpc.DirectionInput, 10, ebcdic("*NONE", 10)},
		{rpc.DirectionInput, 8, make([]byte, 8)}, // legacy exception-mode error code
		{rpc.DirectionInput, 10, ebcdic("*ALL", 10)},
	}
	params := call.Parameters()
	if len(params) != len(want) {
		t.Fatalf("%d parameters", len(params))
	}
	for i, w := range want {
		p := params[i]
		if p.Direction() != w.dir || p.Maximum() != w.max || !bytes.Equal(p.InputBytes(), w.input) {
			t.Errorf("parameter %d = %v % x; want %s max %d % x", i+1, p, p.InputBytes(), w.dir, w.max, w.input)
		}
	}
}

func TestNewCallFormatsAndSelections(t *testing.T) {
	for name, got := range map[string]string{
		"AUTU0100": callFormat(t, qgyolaus.AUTU0100()), "AUTU0150": callFormat(t, qgyolaus.AUTU0150()),
		"AUTU0200": callFormat(t, qgyolaus.AUTU0200()), "AUTU0250": callFormat(t, qgyolaus.AUTU0250()),
	} {
		if got != name {
			t.Errorf("format parameter = %q, want %q", got, name)
		}
	}
	for sel, want := range map[qgyolaus.Selection]string{
		qgyolaus.SelectAll: "*ALL", qgyolaus.SelectUsers: "*USER", qgyolaus.SelectGroups: "*GROUP",
	} {
		c, err := qgyolaus.NewCall(qgyolaus.AUTU0100(), qgyolaus.Request{Selection: sel}, codec)
		if err != nil || !bytes.Equal(c.Parameters()[5].InputBytes(), ebcdic(want, 10)) {
			t.Errorf("selection %s: %v", want, err)
		}
	}
	c, err := qgyolaus.NewCall(qgyolaus.AUTU0100(), qgyolaus.Request{Selection: qgyolaus.SelectMembers, Group: "devgrp", Profile: "AB*"}, codec)
	if err != nil {
		t.Fatal(err)
	}
	p := c.Parameters()
	if !bytes.Equal(p[5].InputBytes(), ebcdic("*MEMBER", 10)) || !bytes.Equal(p[6].InputBytes(), ebcdic("DEVGRP", 10)) || !bytes.Equal(p[8].InputBytes(), ebcdic("AB*", 10)) {
		t.Error("member selection, group or generic profile encoded wrongly")
	}
	lib, _ := rpc.NewObjectName("QSYS")
	if c, _ := qgyolaus.NewCall(qgyolaus.AUTU0100(), qgyolaus.Request{Library: lib}, codec); c.Library().String() != "QSYS" {
		t.Error("library override ignored")
	}
}

func callFormat(t *testing.T, f interface{ Name() string }) string { return f.Name() }

func TestNewCallRejects(t *testing.T) {
	for name, req := range map[string]qgyolaus.Request{
		"members without group": {Selection: qgyolaus.SelectMembers},
		"group without members": {Selection: qgyolaus.SelectAll, Group: "DEV"},
		"bad group":             {Selection: qgyolaus.SelectMembers, Group: "BAD GROUP"},
		"bad profile":           {Profile: "A/B"},
		"double star":           {Profile: "A**"},
		"long generic":          {Profile: "ABCDEFGHIJ*"},
		"undefined selection":   {Selection: 9},
	} {
		if _, err := qgyolaus.NewCall(qgyolaus.AUTU0150(), req, codec); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	var zero qgyolaus.Format[structures.AUTU0150]
	if _, err := qgyolaus.NewCall(zero, qgyolaus.Request{}, codec); err == nil || !zero.IsZero() {
		t.Error("zero Format accepted")
	}
	if _, err := zero.DecodeRecords(nil, 0, codec); err == nil {
		t.Error("zero Format decoded records")
	}
}

func TestOpen(t *testing.T) {
	fc := &fakeCaller{respond: func(c rpc.Call) (*rpc.Result, error) {
		return result(c, listInfo(63, 0, handle, 62, 0xF2), nil), nil
	}}
	res, recovered, err := qgyolaus.Open(context.Background(), fc, qgyolaus.AUTU0150(), qgyolaus.Request{}, codec, codec)
	if err != nil || !recovered || res.Handle != handle || res.Info.TotalRecords != 63 {
		t.Fatalf("Open = %+v, %t, %v", res, recovered, err)
	}
	if len(fc.calls) != 1 {
		t.Errorf("calls = %d", len(fc.calls))
	}
	// Record length 0 is accepted at open (no record returned).
	fc.respond = func(c rpc.Call) (*rpc.Result, error) { return result(c, listInfo(0, 0, handle, 0, 0xF2), nil), nil }
	if _, _, err := qgyolaus.Open(context.Background(), fc, qgyolaus.AUTU0150(), qgyolaus.Request{}, codec, codec); err != nil {
		t.Errorf("empty list = %v", err)
	}
}

func TestOpenMetadataFailuresRecoverHandle(t *testing.T) {
	bad := listInfo(63, 0, handle, 62, 0xF2)
	bad[16] = 0x40 // invalid complete indicator
	for name, info := range map[string][]byte{
		"building":         listInfo(63, 0, handle, 62, 0xF1),
		"wrong length":     listInfo(63, 0, handle, 176, 0xF2),
		"records returned": listInfo(63, 1, handle, 62, 0xF2),
		"malformed":        bad,
	} {
		fc := &fakeCaller{respond: func(c rpc.Call) (*rpc.Result, error) { return result(c, info, nil), nil }}
		res, recovered, err := qgyolaus.Open(context.Background(), fc, qgyolaus.AUTU0150(), qgyolaus.Request{}, codec, codec)
		if !errors.Is(err, qgyolaus.ErrMetadata) || !recovered || res == nil || res.Handle != handle {
			t.Errorf("%s: %+v, %t, %v", name, res, recovered, err)
		}
	}
	fc := &fakeCaller{respond: func(c rpc.Call) (*rpc.Result, error) { return result(c, []byte{1, 2}, nil), nil }}
	if _, recovered, err := qgyolaus.Open(context.Background(), fc, qgyolaus.AUTU0150(), qgyolaus.Request{}, codec, codec); err == nil || recovered {
		t.Errorf("short list information: recovered %t, %v", recovered, err)
	}
}

func TestOpenPropagatesCallerAndAPIErrors(t *testing.T) {
	native := errors.New("server reported CPF2204")
	fc := &fakeCaller{respond: func(rpc.Call) (*rpc.Result, error) { return nil, native }}
	if _, recovered, err := qgyolaus.Open(context.Background(), fc, qgyolaus.AUTU0150(), qgyolaus.Request{}, codec, codec); !errors.Is(err, native) || recovered {
		t.Errorf("caller error = %v, recovered %t", err, recovered)
	}
	errc, _ := errorcode.ERRC0100(32)
	apiErr := make([]byte, 32)
	binary.BigEndian.PutUint32(apiErr[0:], 32)
	binary.BigEndian.PutUint32(apiErr[4:], 16)
	copy(apiErr[8:], ebcdic("CPF3C21", 7)) // format not valid
	fc.respond = func(c rpc.Call) (*rpc.Result, error) { return result(c, listInfo(0, 0, handle, 0, 0xF2), apiErr), nil }
	_, recovered, err := qgyolaus.Open(context.Background(), fc, qgyolaus.AUTU0250(), qgyolaus.Request{ErrorCode: errc}, codec, codec)
	var ae *errorcode.APIError
	if !errors.As(err, &ae) || ae.ExceptionID != "CPF3C21" || ae.Program.String() != "QGYOLAUS" || recovered {
		t.Errorf("API error = %v, recovered %t", err, recovered)
	}
}

func TestDecodeRecords(t *testing.T) {
	rec := append(append(ebcdic("ALICE", 10), 0xF0, 0xF0), ebcdic("Dev", 50)...)
	rec2 := append(append(ebcdic("DEVGRP", 10), 0xF1, 0xF1), ebcdic("Group", 50)...)
	recs, err := qgyolaus.AUTU0150().DecodeRecords(append(rec, rec2...), 2, codec)
	if err != nil || len(recs) != 2 || recs[1].Name != "DEVGRP" || !recs[1].GroupHasMembers || recs[0].Description != "Dev" {
		t.Fatalf("records = %+v, %v", recs, err)
	}
	if _, err := qgyolaus.AUTU0150().DecodeRecords(rec, 2, codec); err == nil {
		t.Error("count beyond the receiver accepted")
	}
}

func TestCompatibilityMatchesStructures(t *testing.T) {
	lengths := map[string]uint32{
		"AUTU0100": qgyolaus.AUTU0100().RecordLength(), "AUTU0150": qgyolaus.AUTU0150().RecordLength(),
		"AUTU0200": qgyolaus.AUTU0200().RecordLength(), "AUTU0250": qgyolaus.AUTU0250().RecordLength(),
	}
	for _, c := range qgyolaus.SupportedFormats() {
		if lengths[c.Format] != c.RecordLength || len(c.Evidence) == 0 {
			t.Errorf("%s: compatibility length %d, descriptor %d", c.Format, c.RecordLength, lengths[c.Format])
		}
		delete(lengths, c.Format)
	}
	if len(lengths) != 0 {
		t.Errorf("formats without a compatibility entry: %v", lengths)
	}
}
