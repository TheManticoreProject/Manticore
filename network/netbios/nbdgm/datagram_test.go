package nbdgm

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
)

// sampleHeader returns the header fields shared by the round-trip tests so each
// test can assert the exact wire offsets against known values.
func sampleHeader() (dgmID uint16, ip net.IP, port uint16) {
	return 0xBEEF, net.IPv4(10, 7, 0, 20), 138
}

func TestDirectUniqueRoundTrip(t *testing.T) {
	dgmID, ip, port := sampleHeader()
	d := &Datagram{
		MsgType:         MsgTypeDirectUnique,
		Flags:           FlagFirst | ((NodeTypeM << sntShift) & sntMask),
		DgmID:           dgmID,
		SourceIP:        ip,
		SourcePort:      port,
		SourceName:      Name{Name: "SENDER", Suffix: 0x00},
		DestinationName: Name{Name: "TARGET", Suffix: 0x20},
		UserData:        []byte("hello datagram"),
	}

	wire, err := d.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Assert the fixed header field offsets on the wire.
	if wire[0] != MsgTypeDirectUnique {
		t.Errorf("MSG_TYPE = 0x%02x, want 0x%02x", wire[0], MsgTypeDirectUnique)
	}
	if wire[1] != FlagFirst|((NodeTypeM<<sntShift)&sntMask) {
		t.Errorf("FLAGS = 0x%02x, want 0x%02x", wire[1], FlagFirst|((NodeTypeM<<sntShift)&sntMask))
	}
	if got := binary.BigEndian.Uint16(wire[2:4]); got != dgmID {
		t.Errorf("DGM_ID = 0x%04x, want 0x%04x", got, dgmID)
	}
	if !bytes.Equal(wire[4:8], ip.To4()) {
		t.Errorf("SOURCE_IP = %v, want %v", net.IP(wire[4:8]), ip.To4())
	}
	if got := binary.BigEndian.Uint16(wire[8:10]); got != port {
		t.Errorf("SOURCE_PORT = %d, want %d", got, port)
	}
	// DGM_LENGTH must equal the trailer length; PACKET_OFFSET is 0.
	trailerLen := len(wire) - directHeaderLen
	if got := binary.BigEndian.Uint16(wire[10:12]); int(got) != trailerLen {
		t.Errorf("DGM_LENGTH = %d, want %d", got, trailerLen)
	}
	if got := binary.BigEndian.Uint16(wire[12:14]); got != 0 {
		t.Errorf("PACKET_OFFSET = %d, want 0", got)
	}
	// Each encoded name is 34 bytes (0x20 + 32 + 0x00) for the default scope.
	if wire[directHeaderLen] != 0x20 {
		t.Errorf("SOURCE_NAME length byte = 0x%02x, want 0x20", wire[directHeaderLen])
	}

	var got Datagram
	n, err := got.Unmarshal(wire)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != len(wire) {
		t.Errorf("Unmarshal consumed %d bytes, want %d", n, len(wire))
	}
	if got.MsgType != d.MsgType || got.DgmID != d.DgmID || got.SourcePort != d.SourcePort {
		t.Errorf("header mismatch: %+v", got)
	}
	if got.NodeType() != NodeTypeM {
		t.Errorf("NodeType = %d, want %d", got.NodeType(), NodeTypeM)
	}
	if !got.SourceIP.Equal(ip) {
		t.Errorf("SOURCE_IP = %v, want %v", got.SourceIP, ip)
	}
	if got.SourceName.Name != "SENDER" || got.SourceName.Suffix != 0x00 {
		t.Errorf("SOURCE_NAME = %+v", got.SourceName)
	}
	if got.DestinationName.Name != "TARGET" || got.DestinationName.Suffix != 0x20 {
		t.Errorf("DESTINATION_NAME = %+v", got.DestinationName)
	}
	if !bytes.Equal(got.UserData, d.UserData) {
		t.Errorf("USER_DATA = %q, want %q", got.UserData, d.UserData)
	}
}

func TestDirectGroupAndBroadcastRoundTrip(t *testing.T) {
	for _, mt := range []uint8{MsgTypeDirectGroup, MsgTypeBroadcast} {
		d := &Datagram{
			MsgType:         mt,
			Flags:           FlagFirst,
			DgmID:           0x0102,
			SourceIP:        net.IPv4(192, 168, 1, 5),
			SourcePort:      138,
			SourceName:      Name{Name: "WK", Suffix: 0x00},
			DestinationName: Name{Name: "WORKGROUP", Suffix: 0x1D},
			UserData:        []byte{0x01, 0x02, 0x03},
		}
		wire, err := d.Marshal()
		if err != nil {
			t.Fatalf("Marshal(0x%02x): %v", mt, err)
		}
		if wire[0] != mt {
			t.Errorf("MSG_TYPE = 0x%02x, want 0x%02x", wire[0], mt)
		}
		var got Datagram
		if _, err := got.Unmarshal(wire); err != nil {
			t.Fatalf("Unmarshal(0x%02x): %v", mt, err)
		}
		if got.DestinationName.Name != "WORKGROUP" || got.DestinationName.Suffix != 0x1D {
			t.Errorf("DESTINATION_NAME = %+v", got.DestinationName)
		}
		if !bytes.Equal(got.UserData, d.UserData) {
			t.Errorf("USER_DATA = %v", got.UserData)
		}
	}
}

func TestDatagramErrorRoundTrip(t *testing.T) {
	d := &Datagram{
		MsgType:    MsgTypeError,
		Flags:      FlagFirst,
		DgmID:      0x2211,
		SourceIP:   net.IPv4(10, 0, 0, 1),
		SourcePort: 138,
		ErrorCode:  ErrorDestinationNameNotPresent,
	}
	wire, err := d.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != baseHeaderLen+1 {
		t.Fatalf("DATAGRAM ERROR length = %d, want %d", len(wire), baseHeaderLen+1)
	}
	if wire[0] != MsgTypeError {
		t.Errorf("MSG_TYPE = 0x%02x, want 0x%02x", wire[0], MsgTypeError)
	}
	if wire[baseHeaderLen] != ErrorDestinationNameNotPresent {
		t.Errorf("ERROR_CODE = 0x%02x, want 0x%02x", wire[baseHeaderLen], ErrorDestinationNameNotPresent)
	}

	var got Datagram
	n, err := got.Unmarshal(wire)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != baseHeaderLen+1 {
		t.Errorf("consumed %d bytes, want %d", n, baseHeaderLen+1)
	}
	if got.ErrorCode != ErrorDestinationNameNotPresent {
		t.Errorf("ERROR_CODE = 0x%02x", got.ErrorCode)
	}
}

func TestQueryRequestAndResponsesRoundTrip(t *testing.T) {
	for _, mt := range []uint8{MsgTypeQueryRequest, MsgTypePositiveQueryResponse, MsgTypeNegativeQueryResponse} {
		d := &Datagram{
			MsgType:         mt,
			Flags:           FlagFirst,
			DgmID:           0x00FF,
			SourceIP:        net.IPv4(172, 16, 0, 9),
			SourcePort:      138,
			DestinationName: Name{Name: "HOST", Suffix: 0x00},
		}
		wire, err := d.Marshal()
		if err != nil {
			t.Fatalf("Marshal(0x%02x): %v", mt, err)
		}
		if wire[0] != mt {
			t.Errorf("MSG_TYPE = 0x%02x, want 0x%02x", wire[0], mt)
		}
		// Base header (10) + a default-scope name (34 bytes).
		if len(wire) != baseHeaderLen+34 {
			t.Errorf("length = %d, want %d", len(wire), baseHeaderLen+34)
		}
		var got Datagram
		n, err := got.Unmarshal(wire)
		if err != nil {
			t.Fatalf("Unmarshal(0x%02x): %v", mt, err)
		}
		if n != len(wire) {
			t.Errorf("consumed %d, want %d", n, len(wire))
		}
		if got.DestinationName.Name != "HOST" {
			t.Errorf("DESTINATION_NAME = %+v", got.DestinationName)
		}
	}
}

func TestNameWithScopeRoundTrip(t *testing.T) {
	n := Name{Name: "SRV", Suffix: 0x20, Scope: "example.com"}
	wire, err := n.encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, consumed, err := decodeName(wire, 0)
	if err != nil {
		t.Fatalf("decodeName: %v", err)
	}
	if consumed != len(wire) {
		t.Errorf("consumed %d, want %d", consumed, len(wire))
	}
	if got.Name != "SRV" || got.Suffix != 0x20 || got.Scope != "example.com" {
		t.Errorf("decoded name = %+v", got)
	}
}

func TestUnmarshalRejectsMalformedInput(t *testing.T) {
	cases := map[string][]byte{
		"empty":                     {},
		"short base header":         {0x10, 0x02, 0x00},
		"direct missing ext header": {MsgTypeDirectUnique, FlagFirst, 0, 1, 10, 0, 0, 1, 0, 138, 0},
		"dgm_length overruns": func() []byte {
			b := make([]byte, directHeaderLen)
			b[0] = MsgTypeDirectUnique
			b[1] = FlagFirst
			binary.BigEndian.PutUint16(b[10:12], 999) // claims 999 trailer bytes, none present
			return b
		}(),
		"error too short": {MsgTypeError, FlagFirst, 0, 1, 10, 0, 0, 1, 0, 138},
		"query truncated name": func() []byte {
			b := make([]byte, baseHeaderLen+1)
			b[0] = MsgTypeQueryRequest
			b[baseHeaderLen] = 0x20 // promises 32 name bytes that are absent
			return b
		}(),
		"unknown msg type": {0x7F, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			var d Datagram
			// Must return an error, never panic.
			if _, err := d.Unmarshal(data); err == nil {
				t.Errorf("Unmarshal(%v) = nil error, want error", data)
			}
		})
	}
}

func TestMarshalRejectsIPv6Source(t *testing.T) {
	d := &Datagram{
		MsgType:         MsgTypeDirectUnique,
		Flags:           FlagFirst,
		SourceIP:        net.ParseIP("fe80::1"),
		DestinationName: Name{Name: "X"},
		SourceName:      Name{Name: "Y"},
	}
	if _, err := d.Marshal(); err == nil {
		t.Error("Marshal with IPv6 SOURCE_IP = nil error, want error")
	}
}
