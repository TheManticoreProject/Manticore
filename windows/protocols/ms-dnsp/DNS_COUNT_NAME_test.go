package msdnsp_test

import (
	"bytes"
	"testing"

	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// TestDNSCountNameWireShape pins the exact on-directory encoding of "example.com":
// 07 'example' 03 'com' 00, with Length 13 (including the null terminator) and LabelCount 2.
func TestDNSCountNameWireShape(t *testing.T) {
	n, err := msdnsp.NewDNS_COUNT_NAMEFromFQDN("example.com")
	if err != nil {
		t.Fatalf("NewDNS_COUNT_NAMEFromFQDN failed: %v", err)
	}

	if n.Length != 13 {
		t.Errorf("Length = %d; want 13", n.Length)
	}
	if n.LabelCount != 2 {
		t.Errorf("LabelCount = %d; want 2", n.LabelCount)
	}

	wantRaw := []byte{0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00}
	if !bytes.Equal(n.RawName, wantRaw) {
		t.Errorf("RawName = % x; want % x", n.RawName, wantRaw)
	}

	marshalled, err := n.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	want := append([]byte{13, 2}, wantRaw...)
	if !bytes.Equal(marshalled, want) {
		t.Errorf("Marshal = % x; want % x", marshalled, want)
	}
}

// TestDNSCountNameRoundTrip round-trips several names through Marshal/Unmarshal and the
// FQDN accessors.
func TestDNSCountNameRoundTrip(t *testing.T) {
	for _, fqdn := range []string{"example.com", "a.b.c.d.example.org", "single", "trailing.dot."} {
		out, err := msdnsp.NewDNS_COUNT_NAMEFromFQDN(fqdn)
		if err != nil {
			t.Fatalf("%q: NewDNS_COUNT_NAMEFromFQDN failed: %v", fqdn, err)
		}
		marshalled, err := out.Marshal()
		if err != nil {
			t.Fatalf("%q: Marshal failed: %v", fqdn, err)
		}

		in := msdnsp.NewDNS_COUNT_NAME()
		read, err := in.Unmarshal(marshalled)
		if err != nil {
			t.Fatalf("%q: Unmarshal failed: %v", fqdn, err)
		}
		if read != len(marshalled) {
			t.Errorf("%q: Unmarshal read %d bytes; want %d", fqdn, read, len(marshalled))
		}

		got, err := in.GetFQDN()
		if err != nil {
			t.Fatalf("%q: GetFQDN failed: %v", fqdn, err)
		}
		// A trailing dot is normalized away on encode.
		want := fqdn
		if want == "trailing.dot." {
			want = "trailing.dot"
		}
		if got != want {
			t.Errorf("round trip of %q = %q; want %q", fqdn, got, want)
		}
	}
}

// TestDNSCountNameEmpty verifies the empty-name encoding: Length 0, LabelCount 0, no RawName.
func TestDNSCountNameEmpty(t *testing.T) {
	n, err := msdnsp.NewDNS_COUNT_NAMEFromFQDN("")
	if err != nil {
		t.Fatalf("NewDNS_COUNT_NAMEFromFQDN(\"\") failed: %v", err)
	}
	if n.Length != 0 || n.LabelCount != 0 || len(n.RawName) != 0 {
		t.Errorf("empty name = {Length:%d LabelCount:%d RawName:% x}; want all zero/empty", n.Length, n.LabelCount, n.RawName)
	}

	marshalled, err := n.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if !bytes.Equal(marshalled, []byte{0, 0}) {
		t.Errorf("Marshal of empty name = % x; want 00 00", marshalled)
	}

	in := msdnsp.NewDNS_COUNT_NAME()
	read, err := in.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != 2 {
		t.Errorf("Unmarshal read %d bytes; want 2", read)
	}
	got, err := in.GetFQDN()
	if err != nil {
		t.Fatalf("GetFQDN failed: %v", err)
	}
	if got != "" {
		t.Errorf("GetFQDN = %q; want empty", got)
	}
}

// TestDNSCountNameRejectsBadInput verifies validation of empty and over-long labels.
func TestDNSCountNameRejectsBadInput(t *testing.T) {
	if _, err := msdnsp.NewDNS_COUNT_NAMEFromFQDN("foo..bar"); err == nil {
		t.Errorf("expected error for empty label, got nil")
	}
	long := make([]byte, 64)
	for i := range long {
		long[i] = 'a'
	}
	if _, err := msdnsp.NewDNS_COUNT_NAMEFromFQDN(string(long)); err == nil {
		t.Errorf("expected error for 64-byte label, got nil")
	}
}

// TestDNSCountNameUnmarshalTruncated verifies that a truncated buffer is rejected rather than
// causing an out-of-bounds read.
func TestDNSCountNameUnmarshalTruncated(t *testing.T) {
	// Length claims 5 bytes of RawName but only 2 are present.
	truncated := []byte{5, 1, 0x03, 'c'}
	in := msdnsp.NewDNS_COUNT_NAME()
	if _, err := in.Unmarshal(truncated); err == nil {
		t.Errorf("expected error for truncated RawName, got nil")
	}
}
