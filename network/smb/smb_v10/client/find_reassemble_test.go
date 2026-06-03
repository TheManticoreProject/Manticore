package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

// makeTrans2ResponseMsg builds a raw SMB_COM_TRANSACTION2 response message carrying
// one fragment: the given total parameter/data sizes, this fragment's parameter and
// data runs, and their displacements. Offsets are header-relative and the payload
// runs are placed immediately after ByteCount (Pad1 = Pad2 = 0).
func makeTrans2ResponseMsg(totalParams int, params []byte, paramDisp int, totalData int, data []byte, dataDisp int) []byte {
	const hdr = 32
	const wordCount = 10
	wordsStart := hdr + 1
	bytesBlockStart := wordsStart + 2*wordCount + 2 // = 55
	paramOffset := bytesBlockStart
	dataOffset := paramOffset + len(params)

	raw := make([]byte, dataOffset+len(data))
	raw[hdr] = byte(wordCount)

	w := raw[wordsStart:]
	binary.LittleEndian.PutUint16(w[0:2], uint16(totalParams))                                           // TotalParameterCount
	binary.LittleEndian.PutUint16(w[2:4], uint16(totalData))                                             // TotalDataCount
	binary.LittleEndian.PutUint16(w[6:8], uint16(len(params)))                                           // ParameterCount
	binary.LittleEndian.PutUint16(w[8:10], uint16(paramOffset))                                          // ParameterOffset
	binary.LittleEndian.PutUint16(w[10:12], uint16(paramDisp))                                           // ParameterDisplacement
	binary.LittleEndian.PutUint16(w[12:14], uint16(len(data)))                                           // DataCount
	binary.LittleEndian.PutUint16(w[14:16], uint16(dataOffset))                                          // DataOffset
	binary.LittleEndian.PutUint16(w[16:18], uint16(dataDisp))                                            // DataDisplacement
	binary.LittleEndian.PutUint16(raw[bytesBlockStart-2:bytesBlockStart], uint16(len(params)+len(data))) // ByteCount

	copy(raw[paramOffset:], params)
	copy(raw[dataOffset:], data)
	return raw
}

func TestReassembleTrans2SingleMessage(t *testing.T) {
	params := []byte{0x01, 0x08}
	data := []byte{0xA0, 0xA1, 0xA2, 0xA3}
	msg := makeTrans2ResponseMsg(len(params), params, 0, len(data), data, 0)

	gotParams, gotData, err := reassembleTrans2(msg, func() ([]byte, error) {
		t.Fatal("recvNext must not be called when the first message is complete")
		return nil, nil
	})
	if err != nil {
		t.Fatalf("reassembleTrans2 error: %v", err)
	}
	if !bytes.Equal(gotParams, params) {
		t.Errorf("params = % x, want % x", gotParams, params)
	}
	if !bytes.Equal(gotData, data) {
		t.Errorf("data = % x, want % x", gotData, data)
	}
}

func TestReassembleTrans2MultiFragment(t *testing.T) {
	// A transaction split across two messages, for both the parameter and data runs.
	frag1 := makeTrans2ResponseMsg(4, []byte{0x11, 0x22}, 0, 6, []byte{0xA0, 0xA1, 0xA2}, 0)
	frag2 := makeTrans2ResponseMsg(4, []byte{0x33, 0x44}, 2, 6, []byte{0xA3, 0xA4, 0xA5}, 3)

	rest := [][]byte{frag2}
	gotParams, gotData, err := reassembleTrans2(frag1, func() ([]byte, error) {
		if len(rest) == 0 {
			return nil, errors.New("recvNext called too many times")
		}
		m := rest[0]
		rest = rest[1:]
		return m, nil
	})
	if err != nil {
		t.Fatalf("reassembleTrans2 error: %v", err)
	}
	if want := []byte{0x11, 0x22, 0x33, 0x44}; !bytes.Equal(gotParams, want) {
		t.Errorf("params = % x, want % x", gotParams, want)
	}
	if want := []byte{0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5}; !bytes.Equal(gotData, want) {
		t.Errorf("data = % x, want % x", gotData, want)
	}
	if len(rest) != 0 {
		t.Errorf("expected exactly one continuation to be consumed")
	}
}

func TestReassembleTrans2OutOfBounds(t *testing.T) {
	// A fragment whose displacement+length exceeds the declared total must be rejected.
	msg := makeTrans2ResponseMsg(0, nil, 0, 8, []byte{0x01, 0x02, 0x03, 0x04}, 6)
	if _, _, err := reassembleTrans2(msg, func() ([]byte, error) {
		return nil, errors.New("should not be reached")
	}); err == nil {
		t.Fatal("expected out-of-bounds fragment error, got nil")
	}
}

func TestReassembleTrans2ReceiveError(t *testing.T) {
	// If a continuation is needed but the transport errors, the error propagates.
	frag1 := makeTrans2ResponseMsg(0, nil, 0, 8, []byte{0xA0, 0xA1, 0xA2, 0xA3}, 0)
	_, _, err := reassembleTrans2(frag1, func() ([]byte, error) {
		return nil, errors.New("connection reset")
	})
	if err == nil {
		t.Fatal("expected receive error to propagate, got nil")
	}
}

func TestParseTrans2Fragment(t *testing.T) {
	params := []byte{0x01, 0x08, 0x05, 0x00}
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	msg := makeTrans2ResponseMsg(10, params, 4, 16, data, 8)

	frag, err := parseTrans2Fragment(msg)
	if err != nil {
		t.Fatalf("parseTrans2Fragment error: %v", err)
	}
	if frag.totalParameterCount != 10 || frag.totalDataCount != 16 {
		t.Errorf("totals = (%d, %d), want (10, 16)", frag.totalParameterCount, frag.totalDataCount)
	}
	if frag.parameterDisplacement != 4 || frag.dataDisplacement != 8 {
		t.Errorf("displacements = (%d, %d), want (4, 8)", frag.parameterDisplacement, frag.dataDisplacement)
	}
	if !bytes.Equal(frag.parameters, params) || !bytes.Equal(frag.data, data) {
		t.Errorf("runs = (% x, % x), want (% x, % x)", frag.parameters, frag.data, params, data)
	}
}
