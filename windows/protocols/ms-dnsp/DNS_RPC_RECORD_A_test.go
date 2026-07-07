package msdnsp_test

import (
	"bytes"
	"net"
	"testing"

	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// TestRecordAWireShape verifies that the IPv4 address is stored as 4 raw bytes in network
// order (192.0.2.1 -> C0 00 02 01).
func TestRecordAWireShape(t *testing.T) {
	r := msdnsp.NewDNS_RPC_RECORD_A()
	if err := r.SetIPv4(net.ParseIP("192.0.2.1")); err != nil {
		t.Fatalf("SetIPv4 failed: %v", err)
	}

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	want := []byte{192, 0, 2, 1}
	if !bytes.Equal(marshalled, want) {
		t.Errorf("Marshal = % x; want % x", marshalled, want)
	}
}

// TestRecordARoundTrip round-trips an A record and checks the recovered address.
func TestRecordARoundTrip(t *testing.T) {
	r := msdnsp.NewDNS_RPC_RECORD_A()
	if err := r.SetIPv4(net.ParseIP("10.20.30.40")); err != nil {
		t.Fatalf("SetIPv4 failed: %v", err)
	}
	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	in := msdnsp.NewDNS_RPC_RECORD_A()
	read, err := in.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != 4 {
		t.Errorf("Unmarshal read %d bytes; want 4", read)
	}
	if got := in.GetIPv4().String(); got != "10.20.30.40" {
		t.Errorf("GetIPv4 = %q; want 10.20.30.40", got)
	}
}

// TestRecordARejectsIPv6 verifies SetIPv4 rejects a non-IPv4 address.
func TestRecordARejectsIPv6(t *testing.T) {
	r := msdnsp.NewDNS_RPC_RECORD_A()
	if err := r.SetIPv4(net.ParseIP("2001:db8::1")); err == nil {
		t.Errorf("expected error setting IPv6 address on A record, got nil")
	}
}

// TestRecordAUnmarshalTruncated verifies that fewer than 4 bytes is an error.
func TestRecordAUnmarshalTruncated(t *testing.T) {
	in := msdnsp.NewDNS_RPC_RECORD_A()
	if _, err := in.Unmarshal([]byte{1, 2, 3}); err == nil {
		t.Errorf("expected error for truncated A record, got nil")
	}
}
