package msdnsp_test

import (
	"bytes"
	"testing"
	"time"

	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// TestRecordTSWireShape verifies EntombedTime is encoded little-endian (unlike the big-endian
// SOA/SRV/preference numeric fields).
func TestRecordTSWireShape(t *testing.T) {
	r := msdnsp.NewDNS_RPC_RECORD_TS()
	r.EntombedTime = 0x0102030405060708

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	// Little-endian: least-significant byte first.
	want := []byte{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
	if !bytes.Equal(marshalled, want) {
		t.Errorf("Marshal = % x; want % x", marshalled, want)
	}
}

// TestRecordTSRoundTrip round-trips the raw EntombedTime value.
func TestRecordTSRoundTrip(t *testing.T) {
	r := msdnsp.NewDNS_RPC_RECORD_TS()
	r.EntombedTime = 133444736000000000

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	in := msdnsp.NewDNS_RPC_RECORD_TS()
	read, err := in.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != 8 {
		t.Errorf("Unmarshal read %d bytes; want 8", read)
	}
	if in.EntombedTime != 133444736000000000 {
		t.Errorf("EntombedTime = %d; want 133444736000000000", in.EntombedTime)
	}
}

// TestRecordTSTimeConversion verifies the FILETIME (100-ns since 1601) conversion helpers by
// round-tripping a whole-second timestamp.
func TestRecordTSTimeConversion(t *testing.T) {
	want := time.Date(2023, time.November, 14, 22, 13, 20, 0, time.UTC)

	r := msdnsp.NewDNS_RPC_RECORD_TS()
	r.SetTime(want)

	got := r.GetTime()
	if !got.Equal(want) {
		t.Errorf("GetTime = %v; want %v", got, want)
	}
}
