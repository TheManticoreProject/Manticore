package types_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestSMB_TIME_DOSMarshalSize(t *testing.T) {
	v := types.NewSMB_TIME_DOSFromTime(13, 30, 14)
	b, err := v.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if len(b) != 2 {
		t.Fatalf("expected 2 bytes, got %d", len(b))
	}
}

func TestSMB_TIME_DOSRoundTrip(t *testing.T) {
	orig := types.NewSMB_TIME_DOSFromTime(13, 30, 14) // 14s -> 7 two-second units
	b, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	// Expected little-endian 16-bit value: hours<<11 | minutes<<5 | twoSeconds
	expected := uint16(13)<<11 | uint16(30)<<5 | uint16(7)
	got := uint16(b[0]) | uint16(b[1])<<8
	if got != expected {
		t.Fatalf("expected packed value 0x%04x, got 0x%04x", expected, got)
	}

	var dec types.SMB_TIME_DOS
	n, err := dec.Unmarshal(b)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 bytes read, got %d", n)
	}
	if dec.Hours != 13 || dec.Minutes != 30 || dec.TwoSeconds != 7 {
		t.Fatalf("round-trip mismatch: %+v", dec)
	}
}

func TestSMB_TIME_DOSUnmarshalShort(t *testing.T) {
	var dec types.SMB_TIME_DOS
	if _, err := dec.Unmarshal([]byte{0x00}); err == nil {
		t.Fatalf("expected error for short input")
	}
}
