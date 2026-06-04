package pdu

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// sampleHeader returns a fully populated header for round-trip and golden tests.
func sampleHeader() Header {
	return Header{
		RPCVersion:         4,
		PacketType:         PacketTypeRequest,
		Flags1:             Flags1Idempotent,
		Flags2:             0,
		DataRepresentation: DataRepresentationLittleEndian,
		SerialHi:           0x00,
		ObjectID:           guid.GUID{A: 0x00010203, B: 0x0405, C: 0x0607, D: 0x0809, E: 0x0a0b0c0d0e0f},
		InterfaceID:        guid.GUID{A: 0x10111213, B: 0x1415, C: 0x1617, D: 0x1819, E: 0x1a1b1c1d1e1f},
		ActivityID:         guid.GUID{A: 0x20212223, B: 0x2425, C: 0x2627, D: 0x2829, E: 0x2a2b2c2d2e2f},
		ServerBoot:         0x30313233,
		InterfaceVersion:   0x34353637,
		SequenceNumber:     0x38393a3b,
		OpNum:              0x3c3d,
		InterfaceHint:      0x3e3f,
		ActivityHint:       0x4041,
		BodyLength:         0,
		FragmentNumber:     0x4243,
		AuthProto:          0x44,
		SerialLo:           0x45,
	}
}

func TestHeaderSizeConstant(t *testing.T) {
	if HeaderSize != 80 {
		t.Fatalf("connectionless common header must be 80 octets, got %d", HeaderSize)
	}
}

func TestHeaderRoundTrip(t *testing.T) {
	in := sampleHeader()
	buf, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(buf) != HeaderSize {
		t.Fatalf("marshalled header is %d bytes, want %d", len(buf), HeaderSize)
	}
	var out Header
	n, err := out.Unmarshal(buf)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != HeaderSize {
		t.Fatalf("Unmarshal consumed %d bytes, want %d", n, HeaderSize)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round-trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
}

// TestHeaderGoldenOffsets asserts the on-the-wire byte layout matches the [C706]
// section 12.6.3.1 field table at every offset that the spec fixes.
func TestHeaderGoldenOffsets(t *testing.T) {
	h := sampleHeader()
	buf, err := h.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	checks := []struct {
		name string
		off  int
		want []byte
	}{
		{"rpc_vers", 0, []byte{0x04}},
		{"ptype", 1, []byte{0x00}},
		{"flags1", 2, []byte{0x20}},
		{"flags2", 3, []byte{0x00}},
		{"drep", 4, []byte{0x10, 0x00, 0x00}},
		{"serial_hi", 7, []byte{0x00}},
		{"object", 8, []byte{0x03, 0x02, 0x01, 0x00, 0x05, 0x04, 0x07, 0x06, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}},
		{"if_id", 24, []byte{0x13, 0x12, 0x11, 0x10, 0x15, 0x14, 0x17, 0x16, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f}},
		{"act_id", 40, []byte{0x23, 0x22, 0x21, 0x20, 0x25, 0x24, 0x27, 0x26, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f}},
		{"server_boot", 56, []byte{0x33, 0x32, 0x31, 0x30}},
		{"if_vers", 60, []byte{0x37, 0x36, 0x35, 0x34}},
		{"seqnum", 64, []byte{0x3b, 0x3a, 0x39, 0x38}},
		{"opnum", 68, []byte{0x3d, 0x3c}},
		{"ihint", 70, []byte{0x3f, 0x3e}},
		{"ahint", 72, []byte{0x41, 0x40}},
		{"len", 74, []byte{0x00, 0x00}},
		{"fragnum", 76, []byte{0x43, 0x42}},
		{"auth_proto", 78, []byte{0x44}},
		{"serial_lo", 79, []byte{0x45}},
	}
	for _, c := range checks {
		got := buf[c.off : c.off+len(c.want)]
		if !bytes.Equal(got, c.want) {
			t.Errorf("field %s at offset %d: got % x, want % x", c.name, c.off, got, c.want)
		}
	}
}

func TestHeaderSerialReassembly(t *testing.T) {
	h := Header{SerialHi: 0xAB, SerialLo: 0xCD}
	if got := h.Serial(); got != 0xABCD {
		t.Fatalf("Serial() = 0x%04x, want 0xABCD", got)
	}
}

func TestNewHeaderDefaults(t *testing.T) {
	h := NewHeader(PacketTypeFack)
	if h.RPCVersion != 4 {
		t.Errorf("RPCVersion = %d, want 4", h.RPCVersion)
	}
	if h.PacketType != PacketTypeFack {
		t.Errorf("PacketType = %s, want fack", h.PacketType)
	}
	if h.InterfaceHint != NoHint || h.ActivityHint != NoHint {
		t.Errorf("hints = %#x/%#x, want both 0xFFFF", h.InterfaceHint, h.ActivityHint)
	}
	if h.DataRepresentation != DataRepresentationLittleEndian {
		t.Errorf("drep = % x, want % x", h.DataRepresentation, DataRepresentationLittleEndian)
	}
}

func TestHeaderRejectsBigEndianDrep(t *testing.T) {
	h := sampleHeader()
	buf, _ := h.Marshal()
	buf[4] = 0x00 // big-endian integer representation
	var out Header
	if _, err := out.Unmarshal(buf); err == nil {
		t.Fatal("expected Unmarshal to reject big-endian drep, got nil error")
	}
}

func TestHeaderTruncated(t *testing.T) {
	var h Header
	if _, err := h.Unmarshal(make([]byte, HeaderSize-1)); err == nil {
		t.Fatal("expected error on truncated header, got nil")
	}
}

func TestPDURoundTrip(t *testing.T) {
	cases := []struct {
		name string
		pt   PacketType
		body []byte
	}{
		{"request", PacketTypeRequest, []byte{0xde, 0xad, 0xbe, 0xef}},
		{"response", PacketTypeResponse, []byte{0x01, 0x02, 0x03}},
		{"ping", PacketTypePing, nil},
		{"ack", PacketTypeAck, nil},
		{"working", PacketTypeWorking, nil},
		{"fault", PacketTypeFault, MarshalStatusBody(0x1c010003)},
		{"reject", PacketTypeReject, MarshalStatusBody(0x1c000008)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := PDU{Header: NewHeader(c.pt), Body: c.body}
			in.Header.SequenceNumber = 42
			wire, err := in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if want := HeaderSize + len(c.body); len(wire) != want {
				t.Fatalf("wire length %d, want %d", len(wire), want)
			}
			var out PDU
			n, err := out.Unmarshal(wire)
			if err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if n != len(wire) {
				t.Fatalf("consumed %d, want %d", n, len(wire))
			}
			if out.Header.BodyLength != uint16(len(c.body)) {
				t.Errorf("BodyLength = %d, want %d", out.Header.BodyLength, len(c.body))
			}
			if !bytes.Equal(out.Body, c.body) {
				t.Errorf("body = % x, want % x", out.Body, c.body)
			}
			if out.Header.PacketType != c.pt || out.Header.SequenceNumber != 42 {
				t.Errorf("header fields not preserved: %+v", out.Header)
			}
		})
	}
}

func TestPDUBodyTruncated(t *testing.T) {
	in := PDU{Header: NewHeader(PacketTypeRequest), Body: []byte{1, 2, 3, 4, 5, 6, 7, 8}}
	wire, _ := in.Marshal()
	var out PDU
	if _, err := out.Unmarshal(wire[:len(wire)-1]); err == nil {
		t.Fatal("expected error when body is shorter than declared len, got nil")
	}
}

func TestPDUPreservesAuthTrailer(t *testing.T) {
	// A non-zero auth_proto means an auth verifier may trail the body. Unmarshal must
	// consume only header+body and leave the trailer for the caller.
	in := PDU{Header: NewHeader(PacketTypeRequest), Body: []byte{0xaa, 0xbb}}
	wire, _ := in.Marshal()
	trailer := []byte{0x99, 0x88, 0x77}
	wire = append(wire, trailer...)
	var out PDU
	n, err := out.Unmarshal(wire)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != HeaderSize+2 {
		t.Fatalf("consumed %d, want %d (header+body only)", n, HeaderSize+2)
	}
	if !bytes.Equal(wire[n:], trailer) {
		t.Errorf("trailer not preserved: got % x, want % x", wire[n:], trailer)
	}
}

func TestFackBodyRoundTrip(t *testing.T) {
	in := FackBody{
		Version:     0,
		Pad1:        0,
		WindowSize:  8,
		MaxTSDU:     0x1234,
		MaxFragSize: 1450,
		SerialNum:   7,
		SelAck:      []uint32{0xffffffff, 0x0000000f},
	}
	buf, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if want := fackFixedSize + 4*len(in.SelAck); len(buf) != want {
		t.Fatalf("fack body length %d, want %d", len(buf), want)
	}
	var out FackBody
	n, err := out.Unmarshal(buf)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != len(buf) {
		t.Fatalf("consumed %d, want %d", n, len(buf))
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round-trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
}

func TestFackBodyEmptySelack(t *testing.T) {
	in := FackBody{WindowSize: 1, MaxFragSize: 1024}
	buf, _ := in.Marshal()
	var out FackBody
	if _, err := out.Unmarshal(buf); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.SelAck) != 0 {
		t.Fatalf("SelAck = %v, want empty", out.SelAck)
	}
}

// TestFackBodySelackGuard ensures a selack_len that overflows the remaining bytes is
// rejected before allocation, matching the hardening applied to the NDR decoder.
func TestFackBodySelackGuard(t *testing.T) {
	buf := make([]byte, fackFixedSize) // declares some selack elements but carries none
	buf[14] = 0xff
	buf[15] = 0xff // selack_len = 65535
	var out FackBody
	if _, err := out.Unmarshal(buf); err == nil {
		t.Fatal("expected error on oversized selack_len, got nil")
	}
}

func TestFackBodyTruncated(t *testing.T) {
	var out FackBody
	if _, err := out.Unmarshal(make([]byte, fackFixedSize-1)); err == nil {
		t.Fatal("expected error on truncated fack body, got nil")
	}
}

func TestCancelBodyRoundTrip(t *testing.T) {
	in := CancelBody{Version: 0, CancelID: 0xcafef00d}
	buf, _ := in.Marshal()
	if len(buf) != CancelBodySize {
		t.Fatalf("cancel body length %d, want %d", len(buf), CancelBodySize)
	}
	var out CancelBody
	n, err := out.Unmarshal(buf)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != CancelBodySize || in != out {
		t.Fatalf("round-trip mismatch: in %+v out %+v (n=%d)", in, out, n)
	}
}

func TestCancelAckBodyRoundTrip(t *testing.T) {
	for _, accepting := range []bool{true, false} {
		in := CancelAckBody{Version: 0, CancelID: 0x11223344, ServerIsAccepting: accepting}
		buf, _ := in.Marshal()
		if len(buf) != CancelAckBodySize {
			t.Fatalf("cancel_ack body length %d, want %d", len(buf), CancelAckBodySize)
		}
		var out CancelAckBody
		if _, err := out.Unmarshal(buf); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if in != out {
			t.Fatalf("round-trip mismatch: in %+v out %+v", in, out)
		}
	}
}

func TestStatusBodyRoundTrip(t *testing.T) {
	const status = uint32(0x1c010003) // nca_op_rng_error, as an example
	buf := MarshalStatusBody(status)
	if len(buf) != StatusBodySize {
		t.Fatalf("status body length %d, want %d", len(buf), StatusBodySize)
	}
	got, err := UnmarshalStatusBody(buf)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != status {
		t.Fatalf("status = 0x%08x, want 0x%08x", got, status)
	}
	if _, err := UnmarshalStatusBody(buf[:2]); err == nil {
		t.Fatal("expected error on truncated status body, got nil")
	}
}

// TestNoCallCarriesFackBody verifies a nocall PDU can carry an optional fack-format
// body and round-trip through the generic framing layer.
func TestNoCallCarriesFackBody(t *testing.T) {
	fb := FackBody{WindowSize: 4, MaxFragSize: 1024, SerialNum: 3, SelAck: []uint32{0x1}}
	body, _ := fb.Marshal()
	in := PDU{Header: NewHeader(PacketTypeNoCall), Body: body}
	wire, _ := in.Marshal()

	var out PDU
	if _, err := out.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal PDU: %v", err)
	}
	var decoded FackBody
	if _, err := decoded.Unmarshal(out.Body); err != nil {
		t.Fatalf("Unmarshal fack body: %v", err)
	}
	if !reflect.DeepEqual(fb, decoded) {
		t.Fatalf("fack body mismatch:\n in: %+v\nout: %+v", fb, decoded)
	}
}
