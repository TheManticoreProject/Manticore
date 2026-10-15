package message

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// unhex decodes a hexdump body (offsets and spaces are ignored).
func unhex(t *testing.T, dump string) []byte {
	t.Helper()
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(dump), "\n") {
		_, rest, _ := strings.Cut(line, ":")
		b.WriteString(strings.ReplaceAll(rest, " ", ""))
	}
	out, err := hex.DecodeString(b.String())
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func marshal(t *testing.T, f hostserver.Frame) []byte {
	t.Helper()
	b, err := f.Marshal(0)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// myQueue is MYQUEUE in QGPL, CCSID 37.
var myQueue = Object{
	Queue:   [10]byte{0xD4, 0xE8, 0xD8, 0xE4, 0xC5, 0xE4, 0xC5, 0x40, 0x40, 0x40},
	Library: [10]byte{0xD8, 0xC7, 0xD7, 0xD3, 0x40, 0x40, 0x40, 0x40, 0x40, 0x40},
}

// The golden vectors are the constructor vectors of iSeriesNavigatorReverse
// docs/16 (correlation ID zero).
func TestGoldenRequests(t *testing.T) {
	read, err := ReadRequest{Object: myQueue}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	write, err := WriteRequest{Object: myQueue, Entry: []byte{1, 2, 3}}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	for name, tt := range map[string]struct {
		f    hostserver.Frame
		want string
	}{
		"exchange attributes": {ExchangeAttributesFrame(0), `
00000000: 00 00 00 1a 00 00 e0 07 00 00 00 00 00 00 00 00
00000010: 00 06 00 00 00 00 00 01 00 00`},
		"query attributes": {QueryAttributesFrame(0, myQueue), `
00000000: 00 00 00 28 00 00 e0 07 00 00 00 00 00 00 00 00
00000010: 00 14 00 01 d4 e8 d8 e4 c5 e4 c5 40 40 40 d8 c7
00000020: d7 d3 40 40 40 40 40 40`},
		"zero-wait destructive read": {read, `
00000000: 00 00 00 30 00 00 e0 07 00 00 00 00 00 00 00 00
00000010: 00 1c 00 02 d4 e8 d8 e4 c5 e4 c5 40 40 40 d8 c7
00000020: d7 d3 40 40 40 40 40 40 f0 00 00 00 00 00 00 f0`},
		"non-keyed write": {write, `
00000000: 00 00 00 33 00 00 e0 07 00 00 00 00 00 00 00 00
00000010: 00 16 00 05 d4 e8 d8 e4 c5 e4 c5 40 40 40 d8 c7
00000020: d7 d3 40 40 40 40 40 40 f0 f1 00 00 00 09 50 01
00000030: 01 02 03`},
	} {
		if got, want := marshal(t, tt.f), unhex(t, tt.want); !bytes.Equal(got, want) {
			t.Errorf("%s:\n got % x\nwant % x", name, got, want)
		}
	}
}

// The remaining requests have no published vector; they are checked field
// by field against the docs/16 layout tables.
func TestLayoutRequests(t *testing.T) {
	object := func(b []byte) {
		t.Helper()
		if !bytes.Equal(b[20:40], append(myQueue.Queue[:], myQueue.Library[:]...)) {
			t.Errorf("object bytes = % x", b[20:40])
		}
	}

	b := marshal(t, DeleteFrame(7, myQueue))
	object(b)
	if len(b) != 40 || b[17] != 20 || b[19] != 0x04 || b[15] != 7 {
		t.Errorf("delete = % x", b)
	}

	f, err := ReadRequest{Correlation: 1, Object: myQueue, Keyed: true, Search: [2]byte{0xC5, 0xD8}, Key: []byte("K1"), Wait: -1, Peek: true}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	b = marshal(t, f)
	object(b)
	want := []byte{0xF1, 0xC5, 0xD8, 0xFF, 0xFF, 0xFF, 0xFF, 0xF1, 0, 0, 0, 8, 0x50, 0x02, 'K', '1'}
	if len(b) != 56 || b[17] != 28 || !bytes.Equal(b[40:], want) {
		t.Errorf("keyed peek = % x", b)
	}

	f, err = WriteRequest{Object: myQueue, Entry: []byte{}, Key: []byte{9}}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	b = marshal(t, f)
	if want := []byte{0xF1, 0xF1, 0, 0, 0, 6, 0x50, 0x01, 0, 0, 0, 7, 0x50, 0x02, 9}; b[17] != 22 || !bytes.Equal(b[40:], want) {
		t.Errorf("keyed write of an empty entry = % x", b)
	}

	f, err = ClearFrame(0, myQueue, nil)
	if err != nil {
		t.Fatal(err)
	}
	if b = marshal(t, f); len(b) != 41 || b[17] != 21 || b[40] != 0xF0 || b[19] != 0x06 {
		t.Errorf("clear all = % x", b)
	}
	f, err = ClearFrame(0, myQueue, []byte("AB"))
	if err != nil {
		t.Fatal(err)
	}
	if b = marshal(t, f); len(b) != 49 || b[17] != 21 || !bytes.Equal(b[40:], []byte{0xF1, 0, 0, 0, 8, 0x50, 0x02, 'A', 'B'}) {
		t.Errorf("clear key = % x", b)
	}

	desc := [50]byte{}
	for i := range desc {
		desc[i] = 0x40
	}
	f, err = CreateRequest{Object: myQueue, MaxEntry: 64512, Authority: AuthorityExclude, SaveSender: true, Type: TypeKeyed, KeyLength: 16, ForceAuxStorage: false, Description: desc}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	b = marshal(t, f)
	object(b)
	if want := []byte{0, 0, 0xFC, 0x00, 0xF2, 0xF1, 0xF2, 0, 16, 0xF0}; len(b) != 100 || b[17] != 80 || b[19] != 0x03 || !bytes.Equal(b[40:50], want) || !bytes.Equal(b[50:], desc[:]) {
		t.Errorf("create = % x", b)
	}
}

func TestRequestValidation(t *testing.T) {
	for name, build := range map[string]func() error{
		"wait below -1":    func() error { _, err := ReadRequest{Wait: -2}.Frame(); return err },
		"empty keyed read": func() error { _, err := ReadRequest{Keyed: true}.Frame(); return err },
		"key on sequential read": func() error {
			_, err := ReadRequest{Key: []byte{1}}.Frame()
			return err
		},
		"long key":        func() error { _, err := WriteRequest{Key: make([]byte, 257)}.Frame(); return err },
		"empty write key": func() error { _, err := WriteRequest{Key: []byte{}}.Frame(); return err },
		"long entry":      func() error { _, err := WriteRequest{Entry: make([]byte, MaxEntryLength+1)}.Frame(); return err },
		"empty clear key": func() error { _, err := ClearFrame(0, Object{}, []byte{}); return err },
		"zero max entry":  func() error { _, err := CreateRequest{Authority: AuthorityAll}.Frame(); return err },
		"authority":       func() error { _, err := CreateRequest{MaxEntry: 1, Authority: 0xF5}.Frame(); return err },
		"type": func() error {
			_, err := CreateRequest{MaxEntry: 1, Authority: AuthorityAll, Type: 3}.Frame()
			return err
		},
		"keyed without key length": func() error {
			_, err := CreateRequest{MaxEntry: 1, Authority: AuthorityAll, Type: TypeKeyed}.Frame()
			return err
		},
		"key length on FIFO": func() error {
			_, err := CreateRequest{MaxEntry: 1, Authority: AuthorityAll, KeyLength: 4}.Frame()
			return err
		},
	} {
		if build() == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := (WriteRequest{Entry: make([]byte, MaxEntryLength), Key: make([]byte, MaxKeyLength)}).Frame(); err != nil {
		t.Errorf("maximum lengths rejected: %v", err)
	}
}

func reply(id uint16, body []byte) hostserver.Frame {
	return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDDataQueue, RequestID: id}, Variable: body}
}

func field(cp uint16, v []byte) []byte {
	b, _ := hostserver.AppendField(nil, cp, v)
	return b
}

func TestParseCommonReply(t *testing.T) {
	r, err := ParseCommonReply(reply(ReplyCommon, []byte{0xF0, 0x00}), 100)
	if err != nil || r.ReturnCode != RCSuccess || r.HasMessage {
		t.Fatalf("success = %+v, %v", r, err)
	}
	msg := []byte{0xC3, 0xD7, 0xC6}
	r, err = ParseCommonReply(reply(ReplyCommon, append([]byte{0xF0, 0x01}, field(0x1234, msg)...)), 100)
	if err != nil || r.ReturnCode != RCCommandCheck || !r.HasMessage || r.MessageCodePoint != 0x1234 || !bytes.Equal(r.Message, msg) {
		t.Fatalf("command check = %+v, %v", r, err)
	}
	for name, b := range map[string][]byte{
		"short":         {0xF0},
		"LL below 6":    {0xF0, 0x01, 0, 0, 0, 5, 0x12, 0x34},
		"LL past end":   {0xF0, 0x01, 0, 0, 0, 9, 0x12, 0x34, 1},
		"partial field": {0xF0, 0x01, 0, 0},
		"two fields":    append(append([]byte{0xF0, 0x01}, field(1, nil)...), field(2, nil)...),
		"too long":      append([]byte{0xF0, 0x01}, field(1, make([]byte, 101))...),
	} {
		if _, err := ParseCommonReply(reply(ReplyCommon, b), 100); !errors.Is(err, hostserver.ErrMalformed) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestParseAttributesReply(t *testing.T) {
	b := make([]byte, 61)
	b[2], b[3], b[4], b[5] = 0, 0, 0x01, 0x00 // 256
	b[6], b[7], b[8], b[9], b[10] = 0xF1, 0xF2, 0, 16, 0xF0
	copy(b[11:], bytes.Repeat([]byte{0x40}, 50))
	r, err := ParseAttributesReply(reply(ReplyQueryAttributes, append(b, 0xEE))) // trailing bytes are tolerated
	if err != nil || r.MaxEntry != 256 || !r.SaveSender || r.Type != TypeKeyed || r.KeyLength != 16 || r.ForceAuxStorage || r.Description[49] != 0x40 {
		t.Fatalf("= %+v, %v", r, err)
	}
	b[7] = 0xF3
	if _, err := ParseAttributesReply(reply(ReplyQueryAttributes, b)); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("type 3 = %v", err)
	}
	if _, err := ParseAttributesReply(reply(ReplyQueryAttributes, b[:60])); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("short = %v", err)
	}
}

func TestParseReadReply(t *testing.T) {
	head := append([]byte{0, 0}, bytes.Repeat([]byte{0x40}, SenderInfoSize)...)
	r, err := ParseReadReply(reply(ReplyRead, append(append([]byte{}, head...), field(CPEntry, nil)...)))
	if err != nil || r.HasSender || len(r.Entry) != 0 || r.Entry == nil || r.HasKey {
		t.Fatalf("empty entry = %+v, %v", r, err)
	}
	sender := append([]byte{}, head...)
	sender[2] = 0xC1
	b := append(append(sender, field(CPKey, []byte("K"))...), field(CPEntry, []byte{1, 2})...)
	r, err = ParseReadReply(reply(ReplyRead, b))
	if err != nil || !r.HasSender || !bytes.Equal(r.Entry, []byte{1, 2}) || !r.HasKey || string(r.Key) != "K" {
		t.Fatalf("keyed = %+v, %v", r, err)
	}
	b[len(b)-1] = 9 // the returned entry is a copy
	if r.Entry[1] != 2 {
		t.Error("entry aliases the frame")
	}
	for name, body := range map[string][]byte{
		"truncated sender": head[:30],
		"no entry":         append(append([]byte{}, head...), field(CPKey, []byte("K"))...),
		"duplicate entry":  append(append(append([]byte{}, head...), field(CPEntry, nil)...), field(CPEntry, nil)...),
		"LL past end":      append(append([]byte{}, head...), 0, 0, 0, 0x20, 0x50, 0x01),
	} {
		if _, err := ParseReadReply(reply(ReplyRead, body)); !errors.Is(err, hostserver.ErrMalformed) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func FuzzParseReplies(f *testing.F) {
	f.Add([]byte{0xF0, 0x01, 0, 0, 0, 6, 0x11, 0x22})
	f.Add(append(append([]byte{0, 0}, bytes.Repeat([]byte{0x40}, 36)...), field(CPEntry, []byte{1})...))
	f.Fuzz(func(t *testing.T, b []byte) {
		for _, id := range []uint16{ReplyCommon, ReplyQueryAttributes, ReplyRead} {
			fr := reply(id, b)
			if r, err := ParseCommonReply(fr, 1<<16); err == nil && r.HasMessage && len(r.Message) > len(b) {
				t.Fatal("message longer than the body")
			}
			_, _ = ParseAttributesReply(fr)
			if r, err := ParseReadReply(fr); err == nil && len(r.Entry)+len(r.Key) > len(b) {
				t.Fatal("fields longer than the body")
			}
		}
	})
}
