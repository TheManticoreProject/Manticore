package message_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
)

// Golden 0x1002 frames from iSeriesNavigatorReverse docs/10 ("Remote-command
// exchange and run command"): command "ABC" (EBCDIC c1c2c3) with the
// option already mapped: 0 (all downgraded below level 7) and 2.
func TestRunCommandGolden(t *testing.T) {
	for option, dump := range map[byte]string{
		0: `00 00 00 1e 00 00 e0 08 00 00 00 00 00 00 00 00 00 01 10 02 00 00 00 00 09 11 01 c1 c2 c3`,
		2: `00 00 00 1e 00 00 e0 08 00 00 00 00 00 00 00 00 00 01 10 02 02 00 00 00 09 11 01 c1 c2 c3`,
	} {
		f, err := message.RunCommandRequest{MessageOption: option, Command: []byte{0xc1, 0xc2, 0xc3}}.Frame()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := marshal(t, f), golden(t, dump); !bytes.Equal(got, want) {
			t.Errorf("option %d:\n% x\nwant\n% x", option, got, want)
		}
	}
	ccsid := uint32(37)
	f, _ := message.RunCommandRequest{Correlation: 4, MessageOption: 3, Command: []byte{0xc1}, CommandCCSID: &ccsid}.Frame()
	want := golden(t, `00 00 00 20 00 00 e0 08 00 00 00 00 00 00 00 04 00 01 10 02 03 00 00 00 0b 11 04 00 00 00 25 c1`)
	if got := marshal(t, f); !bytes.Equal(got, want) {
		t.Errorf("level-10 form:\n% x\nwant\n% x", got, want)
	}
}

// Golden 0x1003 frame from docs/10 ("Program call with all direction
// classes"): TESTPGM in TESTLIB, option 0, level 8, three parameters.
func TestCallProgramGolden(t *testing.T) {
	want := golden(t, `00 00 00 52 00 00 e0 08 00 00 00 00 00 00 00 00
		00 17 10 03 e3 c5 e2 e3 d7 c7 d4 40 40 40 e3 c5
		e2 e3 d3 c9 c2 40 40 40 00 00 03 00 00 00 0e 11
		03 00 00 00 04 00 0b 01 02 00 00 00 0c 11 03 00
		00 00 08 00 16 00 00 00 0d 11 03 00 00 00 04 00
		21 41`)
	f, err := message.CallProgramRequest{
		Program: [10]byte{0xe3, 0xc5, 0xe2, 0xe3, 0xd7, 0xc7, 0xd4, 0x40, 0x40, 0x40},
		Library: [10]byte{0xe3, 0xc5, 0xe2, 0xe3, 0xd3, 0xc9, 0xc2, 0x40, 0x40, 0x40},
		Parameters: []message.RequestParameter{
			{Usage: message.UsageInput, Maximum: 4, Data: message.TrimTrailingZeros([]byte{1, 2, 0, 0})},
			{Usage: message.UsageOutput, Maximum: 8},
			{Usage: message.UsageInputOutputNew, Maximum: 4, Data: message.TrimTrailingZeros([]byte{0x41, 0})},
		},
	}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	if got := marshal(t, f); !bytes.Equal(got, want) {
		t.Errorf("1003 =\n% x\nwant\n% x", got, want)
	}
}

func TestTrimTrailingZeros(t *testing.T) {
	for in, want := range map[string]string{"": "", "\x00\x00": "", "a\x00": "a", "\x00a": "\x00a", "a\x00b\x00\x00": "a\x00b"} {
		if got := string(message.TrimTrailingZeros([]byte(in))); got != want {
			t.Errorf("trim(%q) = %q", in, got)
		}
	}
}

func TestDecodeRLE(t *testing.T) {
	ok := []struct {
		src  []byte
		n    uint32
		want []byte
	}{
		{nil, 0, []byte{}},
		{[]byte{1, 2, 3}, 3, []byte{1, 2, 3}},
		{[]byte{0x1b, 0x1b, 7}, 2, []byte{0x1b, 7}},
		{[]byte{9, 0x1b, 0xAA, 0xBB, 0x00, 0x03}, 7, []byte{9, 0xAA, 0xBB, 0xAA, 0xBB, 0xAA, 0xBB}},
	}
	for _, tc := range ok {
		got, err := message.DecodeRLE(tc.src, tc.n)
		if err != nil || !bytes.Equal(got, tc.want) {
			t.Errorf("DecodeRLE(% x, %d) = % x, %v", tc.src, tc.n, got, err)
		}
	}
	for name, tc := range map[string]struct {
		src []byte
		n   uint32
	}{
		"below length":         {[]byte{1, 2}, 3},
		"above length":         {[]byte{1, 2, 3}, 2},
		"truncated escape":     {[]byte{1, 0x1b}, 2},
		"truncated repeat":     {[]byte{0x1b, 1, 2, 0}, 4},
		"zero count":           {[]byte{0x1b, 1, 2, 0, 0}, 0},
		"bomb past length":     {[]byte{0x1b, 1, 2, 0xFF, 0xFF}, 1024},
		"repeat overshoots":    {[]byte{0x1b, 1, 2, 0, 3}, 5},
		"escaped literal over": {[]byte{0x1b, 0x1b}, 0},
	} {
		if _, err := message.DecodeRLE(tc.src, tc.n); !errors.Is(err, message.ErrRLE) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
}

func envelope(usage uint16, length uint32, payload ...byte) []byte {
	b := binary.BigEndian.AppendUint32(nil, uint32(12+len(payload)))
	b = binary.BigEndian.AppendUint16(b, message.CPParameter)
	b = binary.BigEndian.AppendUint32(b, length)
	b = binary.BigEndian.AppendUint16(b, usage)
	return append(b, payload...)
}

func TestReadParameter(t *testing.T) {
	b := append(envelope(12, 4, 1, 2), 0xEE)
	p, n, err := message.ReadParameter(b)
	if err != nil || n != 14 || p.Usage != 12 || p.Length != 4 || !bytes.Equal(p.Payload, []byte{1, 2}) {
		t.Fatalf("ReadParameter = %+v, %d, %v", p, n, err)
	}
	bad := envelope(12, 4)
	binary.BigEndian.PutUint16(bad[4:], 0x1101)
	for name, b := range map[string][]byte{
		"short":     {0, 0, 0, 12},
		"LL < 12":   {0, 0, 0, 11, 0x11, 0x03, 0, 0, 0, 0, 0, 12},
		"LL > data": append(envelope(12, 4, 1), 0)[:12],
		"wrong CP":  bad,
	} {
		if _, _, err := message.ReadParameter(b); !errors.Is(err, hostserver.ErrMalformed) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
}

// record builds a classic message record (docs/19).
func record(tag uint16, id string, typ [2]byte, sev uint16, sub, text []byte, ext ...byte) []byte {
	ll := 41 + len(sub) + len(text) + len(ext)
	r := binary.BigEndian.AppendUint32(nil, uint32(ll))
	r = binary.BigEndian.AppendUint16(r, tag)
	idb := make([]byte, 7)
	for i := range idb {
		idb[i] = 0x40
	}
	copy(idb, id)
	r = append(r, idb...)
	r = append(r, typ[:]...)
	r = binary.BigEndian.AppendUint16(r, sev)
	r = append(r, bytes.Repeat([]byte{0x40}, 20)...)
	r = binary.BigEndian.AppendUint16(r, uint16(len(sub)))
	r = binary.BigEndian.AppendUint16(r, uint16(len(text)))
	r = append(append(append(r, sub...), text...), ext...)
	return r
}

var limits = message.MessageLimits{MaxMessages: 8, MaxText: 64}

func TestParseMessages(t *testing.T) {
	b := append(record(0x0103, "CPF9801", [2]byte{0xF1, 0xF5}, 40, []byte{1, 2}, []byte{0xC1, 0x40}, 0xEE),
		record(0, "CPC2191", [2]byte{0xF0, 0xF1}, 0, nil, nil)...)
	msgs, err := message.ParseMessages(2, b, limits)
	if err != nil {
		t.Fatal(err)
	}
	m := msgs[0]
	if m.Tag != 0x0103 || string(m.ID[:]) != "CPF9801" || m.Type != 15 || m.Severity != 40 ||
		!bytes.Equal(m.Substitution, []byte{1, 2}) || !bytes.Equal(m.Text, []byte{0xC1, 0x40}) || !bytes.Equal(m.Extension, []byte{0xEE}) {
		t.Errorf("record 0 = %+v", m)
	}
	if msgs[1].Type != 1 || len(msgs[1].Text) != 0 {
		t.Errorf("record 1 = %+v", msgs[1])
	}
	b[41] = 9 // records are copies
	if msgs[0].Substitution[0] != 1 {
		t.Error("record aliases the reply buffer")
	}
	if msgs, err := message.ParseMessages(0, nil, limits); err != nil || len(msgs) != 0 {
		t.Errorf("zero messages = %v, %v", msgs, err)
	}
}

func TestParseMessagesRejects(t *testing.T) {
	good := record(0, "CPF0001", [2]byte{0xF0, 0xF2}, 10, nil, []byte{0xC1})
	mutate := func(f func([]byte)) []byte {
		b := append([]byte(nil), good...)
		f(b)
		return b
	}
	for name, tc := range map[string]struct {
		count uint16
		b     []byte
	}{
		"count over limit":     {9, good},
		"missing record":       {2, good},
		"40-byte remainder":    {1, good[:40]},
		"length below 41":      {1, mutate(func(b []byte) { binary.BigEndian.PutUint32(b, 40) })},
		"length zero":          {1, mutate(func(b []byte) { binary.BigEndian.PutUint32(b, 0) })},
		"length overrun":       {1, mutate(func(b []byte) { binary.BigEndian.PutUint32(b, 99) })},
		"fields exceed record": {1, mutate(func(b []byte) { binary.BigEndian.PutUint16(b[39:], 5) })},
		"sum overflow":         {1, mutate(func(b []byte) { binary.BigEndian.PutUint16(b[37:], 0xFFFF); binary.BigEndian.PutUint16(b[39:], 0xFFFF) })},
		"text over limit":      {1, record(0, "X", [2]byte{0xF0, 0xF1}, 0, nil, bytes.Repeat([]byte{0xC1}, 65))},
		"bad type nibble":      {1, mutate(func(b []byte) { b[13] = 0xFA })},
		"trailing bytes":       {1, append(append([]byte(nil), good...), 0)},
	} {
		if _, err := message.ParseMessages(tc.count, tc.b, limits); !errors.Is(err, hostserver.ErrMalformed) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
	// Exactly 41 bytes is the minimum valid record.
	if _, err := message.ParseMessages(1, record(0, "A", [2]byte{0xF0, 0xF0}, 0, nil, nil), limits); err != nil {
		t.Errorf("41-byte record rejected: %v", err)
	}
}

func TestParseOperationReply(t *testing.T) {
	f := hostserver.Frame{Header: hostserver.Header{RequestID: 0x9003}, Template: []byte{0x05, 0x00}, Variable: []byte{0, 2, 9}}
	r, err := message.ParseOperationReply(f)
	if err != nil || r.ReplyID != 0x9003 || r.ReturnCode != 0x0500 || r.MessageCount != 2 || !bytes.Equal(r.Rest, []byte{9}) {
		t.Fatalf("reply = %+v, %v", r, err)
	}
	if _, err := message.ParseOperationReply(hostserver.Frame{Variable: []byte{0, 0, 0}}); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("short reply error = %v", err)
	}
}

func FuzzDecodeRLE(f *testing.F) {
	f.Add([]byte{9, 0x1b, 0xAA, 0xBB, 0x00, 0x03}, uint16(7))
	f.Add([]byte{0x1b, 0x1b}, uint16(1))
	f.Add([]byte{0x1b, 1, 2, 0xFF, 0xFF}, uint16(64))
	f.Fuzz(func(t *testing.T, src []byte, n uint16) {
		out, err := message.DecodeRLE(src, uint32(n))
		if err == nil && len(out) != int(n) {
			t.Fatalf("accepted %d bytes for length %d", len(out), n)
		}
	})
}

func FuzzParseMessages(f *testing.F) {
	f.Add(uint16(1), record(0, "CPF9801", [2]byte{0xF1, 0xF5}, 40, []byte{1}, []byte{0xC1}))
	f.Add(uint16(2), append(record(0, "A", [2]byte{0xF0, 0xF0}, 0, nil, nil), record(0, "B", [2]byte{0xF0, 0xF0}, 0, nil, nil)...))
	f.Fuzz(func(t *testing.T, count uint16, b []byte) {
		msgs, err := message.ParseMessages(count, b, limits)
		if err != nil {
			return
		}
		if len(msgs) != int(count) || len(msgs) > limits.MaxMessages {
			t.Fatalf("accepted %d records for count %d", len(msgs), count)
		}
		for _, m := range msgs {
			if m.Type > 99 || uint32(len(m.Text)) > limits.MaxText {
				t.Fatalf("record outside bounds: %+v", m)
			}
		}
	})
}
