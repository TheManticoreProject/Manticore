package errorcode_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/errorcode"
)

var (
	lib, _   = rpc.NewObjectName("QGY")
	pgm, _   = rpc.NewObjectName("QGYOLAUS")
	codec, _ = ccsid.Lookup(ccsid.CCSID37)
)

func ebcdic(s string) []byte {
	b, err := codec.Encode(s)
	if err != nil {
		panic(err)
	}
	return b
}

func TestExceptionMode(t *testing.T) {
	p := errorcode.Exception()
	if p != (errorcode.Parameter{}) || p.Format().String() != "exception" {
		t.Error("the zero Parameter must be exception mode")
	}
	r := p.RPC()
	// JTOpen 4.8 legacy form: input-only, eight zero bytes.
	if r.Direction() != rpc.DirectionInput || r.Maximum() != 8 || !bytes.Equal(r.InputBytes(), make([]byte, 8)) {
		t.Errorf("exception parameter = %v", r)
	}
	if err := p.Check(lib, pgm, rpc.ParameterResult{}, codec); err != nil {
		t.Errorf("exception-mode Check = %v", err)
	}
}

func TestBufferModeParameters(t *testing.T) {
	p, err := errorcode.ERRC0100(64)
	if err != nil {
		t.Fatal(err)
	}
	r := p.RPC()
	if r.Direction() != rpc.DirectionInputOutput || r.Maximum() != 64 || !bytes.Equal(r.InputBytes(), []byte{0, 0, 0, 64}) {
		t.Errorf("ERRC0100 parameter = %v % x", r, r.InputBytes())
	}
	p2, _ := errorcode.ERRC0200(96)
	r2 := p2.RPC()
	if r2.Maximum() != 96 || !bytes.Equal(r2.InputBytes(), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 96}) {
		t.Errorf("ERRC0200 parameter = % x", r2.InputBytes())
	}
	for _, size := range []uint32{0, 8, 15, errorcode.MaxSize + 1} {
		if _, err := errorcode.ERRC0100(size); err == nil {
			t.Errorf("ERRC0100(%d) accepted", size)
		}
	}
	for _, size := range []uint32{12, 31, errorcode.MaxSize + 1} {
		if _, err := errorcode.ERRC0200(size); err == nil {
			t.Errorf("ERRC0200(%d) accepted", size)
		}
	}
}

func errc0100(size uint32, available uint32, id string, data []byte) []byte {
	b := make([]byte, size)
	binary.BigEndian.PutUint32(b[0:], size)
	binary.BigEndian.PutUint32(b[4:], available)
	if id != "" {
		copy(b[8:], ebcdic(id))
		copy(b[16:], data)
	}
	return b
}

func TestCheckERRC0100(t *testing.T) {
	p, _ := errorcode.ERRC0100(32)
	if err := p.Check(lib, pgm, rpc.ParameterResult{Present: true, Data: errc0100(32, 0, "", nil)}, codec); err != nil {
		t.Errorf("no error reported: %v", err)
	}
	err := p.Check(lib, pgm, rpc.ParameterResult{Present: true, Data: errc0100(32, 16+5, "CPF2204", []byte("SECRT"))}, codec)
	var ae *errorcode.APIError
	if !errors.Is(err, errorcode.ErrAPI) || !errors.As(err, &ae) || ae.ExceptionID != "CPF2204" || string(ae.ExceptionData) != "SECRT" || ae.BytesAvailable != 21 {
		t.Fatalf("Check = %#v, %v", ae, err)
	}
	if strings.Contains(err.Error(), "SECRT") || !strings.Contains(err.Error(), "QGY/QGYOLAUS") {
		t.Errorf("error text %q", err)
	}
	// More data available than the buffer holds: truncated to the buffer.
	err = p.Check(lib, pgm, rpc.ParameterResult{Present: true, Data: errc0100(32, 500, "CPF2204", bytes.Repeat([]byte{7}, 16))}, codec)
	if !errors.As(err, &ae) || len(ae.ExceptionData) != 16 {
		t.Errorf("truncated data = %d bytes, %v", len(ae.ExceptionData), err)
	}
}

func TestCheckERRC0200(t *testing.T) {
	p, _ := errorcode.ERRC0200(48)
	b := make([]byte, 48)
	binary.BigEndian.PutUint32(b[0:], 0xFFFFFFFF)
	binary.BigEndian.PutUint32(b[4:], 48)
	binary.BigEndian.PutUint32(b[8:], 36)
	copy(b[12:], ebcdic("CPF9801"))
	binary.BigEndian.PutUint32(b[20:], 37)
	binary.BigEndian.PutUint32(b[24:], 32)
	binary.BigEndian.PutUint32(b[28:], 4)
	copy(b[32:], []byte{1, 2, 3, 4})
	err := p.Check(lib, pgm, rpc.ParameterResult{Present: true, Data: b}, codec)
	var ae *errorcode.APIError
	if !errors.As(err, &ae) || ae.ExceptionID != "CPF9801" || ae.CCSID != 37 || !bytes.Equal(ae.ExceptionData, []byte{1, 2, 3, 4}) || ae.Format != errorcode.FormatERRC0200 {
		t.Fatalf("Check = %#v, %v", ae, err)
	}
	inside := append([]byte(nil), b...)
	binary.BigEndian.PutUint32(inside[24:], 20)
	if err := p.Check(lib, pgm, rpc.ParameterResult{Present: true, Data: inside}, codec); !errors.Is(err, errorcode.ErrMalformed) {
		t.Errorf("offset inside fixed fields = %v", err)
	}
	past := append([]byte(nil), b...)
	binary.BigEndian.PutUint32(past[24:], 1000)
	if err := p.Check(lib, pgm, rpc.ParameterResult{Present: true, Data: past}, codec); !errors.As(err, &ae) || len(ae.ExceptionData) != 0 {
		t.Errorf("offset past the buffer = %v", err)
	}
}

func TestCheckRejectsMalformedSlots(t *testing.T) {
	p, _ := errorcode.ERRC0100(32)
	for name, slot := range map[string]rpc.ParameterResult{
		"absent":           {},
		"short":            {Present: true, Data: make([]byte, 31)},
		"bytes provided":   {Present: true, Data: errc0100(32, 0, "", nil)[:0:0]},
		"provided changed": {Present: true, Data: func() []byte { b := errc0100(32, 0, "", nil); b[3] = 9; return b }()},
	} {
		if err := p.Check(lib, pgm, slot, codec); !errors.Is(err, errorcode.ErrMalformed) {
			t.Errorf("%s: Check = %v", name, err)
		}
	}
}
