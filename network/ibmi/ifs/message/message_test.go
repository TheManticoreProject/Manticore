package message

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

func marshal(t *testing.T, f hostserver.Frame, err error) []byte {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.Marshal(0)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func unhex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.Join(strings.Fields(s), ""))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// utf16be encodes ASCII as UTF-16BE code units.
func utf16be(s string) []byte {
	var b []byte
	for _, r := range s {
		b = append(b, 0, byte(r))
	}
	return b
}

// The vectors come from the installed JTOpen 4.8 constructors:
// iSeriesNavigatorReverse .research/family-av/flag-read-vectors.txt (with
// correlations 1-4) and docs/10 "IFS V5R4 level-8 requests" (correlation 0).
func TestGoldenRequests(t *testing.T) {
	read := func(corr, handle uint32, off int64, n uint32) []byte {
		f, err := ReadFrame(corr, handle, off, n)
		return marshal(t, f, err)
	}
	open := func(corr uint32, dataCCSID uint16, path string) []byte {
		f, err := OpenRequest{Correlation: corr, PathCCSID: 1200, DataCCSID: dataCCSID, Access: AccessRead, Deny: DenyNone, Conversion: ConvertNone, Duplicate: OpenExistingOnly, Path: utf16be(path)}.Frame()
		return marshal(t, f, err)
	}
	exchange := func(corr uint32) []byte {
		f, err := ExchangeAttributesFrame(corr, DataStreamLevel, PreferredCCSIDs)
		return marshal(t, f, err)
	}
	for name, tt := range map[string]struct{ got, want []byte }{
		"exchange-level8": {exchange(1), unhex(t, "0000002a0000e0020000000000000001000a0016000000080006ffffffff0000000c000a04b034b0f200")},
		"ifs-0016-v5r4":   {exchange(0), unhex(t, "0000002a0000e0020000000000000000000a0016000000080006ffffffff0000000c000a04b034b0f200")},
		"open-flag": {open(2, 1200, "/flag"), unhex(t, "000000480000e002000000000000000200240002000004b00000000104b00001000000000008"+
			"000000000000000000010000000000000000000000100002002f0066006c00610067")},
		"ifs-0002-open-read": {open(0, 0xFFFF, "/tmp/a"), unhex(t, `
			00 00 00 4a 00 00 e0 02 00 00 00 00 00 00 00 00
			00 24 00 02 00 00 04 b0 00 00 00 01 ff ff 00 01
			00 00 00 00 00 08 00 00 00 00 00 00 00 00 00 01
			00 00 00 00 00 00 00 00 00 00 00 12 00 02 00 2f
			00 74 00 6d 00 70 00 2f 00 61`)},
		"read-flag":     {read(3, 0xDEADBEEF, 0, 4096), unhex(t, "0000002a0000e0020000000000000003001600030000deadbeef00000000000000000000100000000000")},
		"ifs-0003-read": {read(0, 0x01020304, 16, 32), unhex(t, "0000002a0000e00200000000000000000016000300000102030400000000000000100000002000000000")},
		"close-flag":    {marshal(t, CloseFrame(4, 0xDEADBEEF), nil), unhex(t, "000000290000e0020000000000000004001500090000deadbeef0002ffff0064000000000000000000")},
	} {
		if !bytes.Equal(tt.got, tt.want) {
			t.Errorf("%s:\n got % x\nwant % x", name, tt.got, tt.want)
		}
	}
}

func TestRequestValidation(t *testing.T) {
	if _, err := ExchangeAttributesFrame(1, 8, nil); err == nil {
		t.Error("an empty CCSID list was accepted")
	}
	if _, err := (OpenRequest{}).Frame(); err == nil {
		t.Error("an empty path was accepted")
	}
	for _, tt := range []struct {
		off int64
		n   uint32
	}{{-1, 1}, {MaxOffset + 1, 1}, {0, 0}} {
		if _, err := ReadFrame(1, 1, tt.off, tt.n); err == nil {
			t.Errorf("read at %d for %d accepted", tt.off, tt.n)
		}
	}
	if _, err := ReadFrame(1, 1, MaxOffset, 1); err != nil {
		t.Errorf("the maximum offset was rejected: %v", err)
	}
}

func reply(id uint16, body []byte) hostserver.Frame {
	return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDFile, RequestID: id}, Variable: body}
}

func TestParseAttributesReply(t *testing.T) {
	b := []byte{0, 0, 0, 8, 0, 0, 0, 0, 0xFF, 0xFF, 0, 0, 0, 10, 0, 0x0A, 0x04, 0xB0, 0x34, 0xB0}
	r, err := ParseAttributesReply(reply(ReplyExchangeAttributes, b))
	if err != nil || r.Level != 8 || r.MaxBlock != 0xFFFF || len(r.CCSIDs) != 2 || r.CCSIDs[0] != 1200 || r.CCSIDs[1] != 13488 {
		t.Fatalf("= %+v, %v", r, err)
	}
	for name, body := range map[string][]byte{
		"short":       b[:17],
		"LL below 8":  append(append([]byte{}, b[:10]...), 0, 0, 0, 7, 0, 0x0A, 0x04, 0xB0),
		"LL past end": append(append([]byte{}, b[:10]...), 0, 0, 0, 12, 0, 0x0A, 0x04, 0xB0),
	} {
		if _, err := ParseAttributesReply(reply(ReplyExchangeAttributes, body)); !errors.Is(err, hostserver.ErrMalformed) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestParseOpenReply(t *testing.T) {
	b := make([]byte, OpenReplyLength)
	binary.BigEndian.PutUint32(b[2:], 0x11223344)
	binary.BigEndian.PutUint64(b[6:], 99)
	binary.BigEndian.PutUint16(b[14:], 819)
	binary.BigEndian.PutUint32(b[26:], 1700000000)
	binary.BigEndian.PutUint32(b[30:], 250000)
	binary.BigEndian.PutUint32(b[42:], 0xFFFFFFF0)
	r, err := ParseOpenReply(reply(ReplyOpen, b))
	if err != nil || r.Handle != 0x11223344 || r.FileID != 99 || r.DataCCSID != 819 || r.Size != 0xFFFFFFF0 ||
		!r.Modified.Equal(time.Unix(1700000000, 250000000)) {
		t.Fatalf("= %+v, %v", r, err)
	}
	if _, err := ParseOpenReply(reply(ReplyOpen, b[:OpenReplyLength-1])); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("short = %v", err)
	}
}

func TestParseReadReply(t *testing.T) {
	field := func(chain uint16, data []byte) []byte {
		b := binary.BigEndian.AppendUint16(nil, chain)
		b = binary.BigEndian.AppendUint16(b, 1200)
		b = binary.BigEndian.AppendUint32(b, uint32(6+len(data)))
		b = binary.BigEndian.AppendUint16(b, 0x0020)
		return append(b, data...)
	}
	r, err := ParseReadReply(reply(ReplyRead, field(1, []byte("abc"))))
	if err != nil || !r.Chained || r.DataCCSID != 1200 || r.CodePoint != 0x0020 || string(r.Data) != "abc" {
		t.Fatalf("= %+v, %v", r, err)
	}
	if r, err := ParseReadReply(reply(ReplyRead, field(2, nil))); err != nil || r.Chained || len(r.Data) != 0 {
		t.Errorf("EOF packet (chain bit clear, other bits set) = %+v, %v", r, err)
	}
	bad := field(0, []byte("abc"))
	for name, body := range map[string][]byte{
		"short":       bad[:9],
		"LL past end": append(append([]byte{}, bad[:4]...), 0, 0, 0, 10, 0, 0x20, 'a'),
		"LL below 6":  append(append([]byte{}, bad[:4]...), 0, 0, 0, 5, 0, 0x20),
		"trailing":    append(append([]byte{}, bad...), 0),
	} {
		if _, err := ParseReadReply(reply(ReplyRead, body)); !errors.Is(err, hostserver.ErrMalformed) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if rc, err := ParseReturnCode(reply(ReplyReturnCode, []byte{0, 0, 0, 22})); err != nil || rc != 22 {
		t.Errorf("return code = %d, %v", rc, err)
	}
	if _, err := ParseReturnCode(reply(ReplyClose, []byte{0, 0, 0})); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("short return code = %v", err)
	}
	if !Chained(reply(ReplyRead, []byte{0, 1})) || Chained(reply(ReplyRead, []byte{1})) {
		t.Error("Chained misreads the indicator")
	}
}

func FuzzParseReplies(f *testing.F) {
	f.Add([]byte{0, 0, 0, 8, 0, 0, 0, 0, 0xFF, 0xFF, 0, 0, 0, 8, 0, 0x0A, 0x04, 0xB0})
	f.Add([]byte{0, 1, 0x04, 0xB0, 0, 0, 0, 7, 0, 0x20, 'x'})
	f.Fuzz(func(t *testing.T, b []byte) {
		if r, err := ParseReadReply(reply(ReplyRead, b)); err == nil && len(r.Data) > len(b) {
			t.Fatal("data longer than the body")
		}
		if r, err := ParseAttributesReply(reply(ReplyExchangeAttributes, b)); err == nil && 2*len(r.CCSIDs) > len(b) {
			t.Fatal("more CCSIDs than bytes")
		}
		_, _ = ParseOpenReply(reply(ReplyOpen, b))
		_, _ = ParseReturnCode(reply(ReplyReturnCode, b))
	})
}
