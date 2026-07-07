package dnsrecord_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dns/dnsrecord"
)

// TestRecordSRVWireShape verifies the big-endian encoding of wPriority/wWeight/wPort followed
// by the DNS_COUNT_NAME target.
func TestRecordSRVWireShape(t *testing.T) {
	r := dnsrecord.NewDNS_RPC_RECORD_SRV()
	r.WPriority = 0x0102
	r.WWeight = 0x0304
	r.WPort = 0x0050 // port 80
	if err := r.NameTarget.SetFQDN("srv.example.com"); err != nil {
		t.Fatalf("SetFQDN failed: %v", err)
	}

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	// Big-endian words: 01 02 | 03 04 | 00 50.
	wantHeader := []byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x50}
	if !bytes.Equal(marshalled[:6], wantHeader) {
		t.Errorf("SRV numeric header = % x; want % x", marshalled[:6], wantHeader)
	}
}

// TestRecordSRVRoundTrip round-trips an SRV record and checks every field.
func TestRecordSRVRoundTrip(t *testing.T) {
	r := dnsrecord.NewDNS_RPC_RECORD_SRV()
	r.WPriority = 10
	r.WWeight = 20
	r.WPort = 389
	if err := r.NameTarget.SetFQDN("dc01.example.com"); err != nil {
		t.Fatalf("SetFQDN failed: %v", err)
	}

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	in := dnsrecord.NewDNS_RPC_RECORD_SRV()
	read, err := in.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != len(marshalled) {
		t.Errorf("Unmarshal read %d bytes; want %d", read, len(marshalled))
	}
	if in.WPriority != 10 || in.WWeight != 20 || in.WPort != 389 {
		t.Errorf("numeric fields mismatch: priority=%d weight=%d port=%d", in.WPriority, in.WWeight, in.WPort)
	}
	target, _ := in.NameTarget.GetFQDN()
	if target != "dc01.example.com" {
		t.Errorf("NameTarget = %q; want dc01.example.com", target)
	}
}
