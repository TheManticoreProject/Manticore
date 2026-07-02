package msfrs1

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestCOMM_PACKET_RoundTrip exercises COMM_PACKET: six leading DWORDs, a [unique]
// pointer to a size_is(PktLen) byte buffer, and the two [ignore] void* fields that
// must marshal as NULL referent ids (four zero octets each) and unmarshal back to nil.
func TestCOMM_PACKET_RoundTrip(t *testing.T) {
	in := COMM_PACKET{
		Major:  0,
		Minor:  0x00000009, // NTFRS_COMM_MINOR_9
		CsId:   1,
		MemLen: 8,
		UpkLen: 0,
		Pkt:    []uint8{0x01, 0x00, 0x02, 0x00, 0xde, 0xad, 0xbe, 0xef},
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out COMM_PACKET
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.PktLen != ndr.DWORD(len(in.Pkt)) {
		t.Errorf("PktLen = %d, want %d (count must be derived from Pkt)", out.PktLen, len(in.Pkt))
	}
	if !reflect.DeepEqual(out.Pkt, in.Pkt) {
		t.Errorf("Pkt round-trip: got %v want %v", out.Pkt, in.Pkt)
	}
	if out.Minor != in.Minor || out.CsId != in.CsId || out.MemLen != in.MemLen {
		t.Errorf("scalar fields mismatch: %+v", out)
	}
	if out.DataName != nil || out.DataHandle != nil {
		t.Errorf("DataName/DataHandle must round-trip to nil, got %v/%v", out.DataName, out.DataHandle)
	}
}

// TestCOMM_PACKET_IgnoreFieldsAreNullReferents pins the wire form of the [ignore]
// fields: with a nil Pkt the trailing bytes are Pkt's NULL referent id followed by the
// two [ignore] NULL referent ids — twelve zero octets after the six DWORDs.
func TestCOMM_PACKET_IgnoreFieldsAreNullReferents(t *testing.T) {
	raw, err := ndr.Marshal(&COMM_PACKET{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// 6 DWORDs (24) + Pkt referent id (4) + DataName referent id (4) + DataHandle
	// referent id (4) = 36 octets, all zero when everything is nil/0.
	if len(raw) != 36 {
		t.Fatalf("marshalled length = %d, want 36", len(raw))
	}
	for i, b := range raw {
		if b != 0 {
			t.Fatalf("octet %d = 0x%02x, want 0x00 (all-nil COMM_PACKET must be zero)", i, b)
		}
	}
}
