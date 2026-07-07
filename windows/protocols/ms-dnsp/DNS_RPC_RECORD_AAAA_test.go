package msdnsp_test

import (
	"bytes"
	"net"
	"testing"

	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// TestRecordAAAAWireShape verifies that the IPv6 address is stored as 16 raw bytes in network
// order.
func TestRecordAAAAWireShape(t *testing.T) {
	r := msdnsp.NewDNS_RPC_RECORD_AAAA()
	if err := r.SetIPv6(net.ParseIP("2001:db8::1")); err != nil {
		t.Fatalf("SetIPv6 failed: %v", err)
	}

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(marshalled) != 16 {
		t.Fatalf("Marshal produced %d bytes; want 16", len(marshalled))
	}
	want := []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}
	if !bytes.Equal(marshalled, want) {
		t.Errorf("Marshal = % x; want % x", marshalled, want)
	}
}

// TestRecordAAAARoundTrip round-trips an AAAA record and checks the recovered address.
func TestRecordAAAARoundTrip(t *testing.T) {
	r := msdnsp.NewDNS_RPC_RECORD_AAAA()
	if err := r.SetIPv6(net.ParseIP("fe80::20c:29ff:fe1a:2b3c")); err != nil {
		t.Fatalf("SetIPv6 failed: %v", err)
	}
	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	in := msdnsp.NewDNS_RPC_RECORD_AAAA()
	read, err := in.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != 16 {
		t.Errorf("Unmarshal read %d bytes; want 16", read)
	}
	if got := in.GetIPv6().String(); got != "fe80::20c:29ff:fe1a:2b3c" {
		t.Errorf("GetIPv6 = %q; want fe80::20c:29ff:fe1a:2b3c", got)
	}
}

// TestRecordAAAAUnmarshalTruncated verifies that fewer than 16 bytes is an error.
func TestRecordAAAAUnmarshalTruncated(t *testing.T) {
	in := msdnsp.NewDNS_RPC_RECORD_AAAA()
	if _, err := in.Unmarshal(make([]byte, 15)); err == nil {
		t.Errorf("expected error for truncated AAAA record, got nil")
	}
}
