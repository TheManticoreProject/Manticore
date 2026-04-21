package kerberos

import (
	"bytes"
	"testing"
)

// TestEncodeKDCOptionsBitLayout verifies the bit-position-to-byte mapping
// used by encodeKDCOptions against RFC 4120 §5.2.8 / §5.4.1. Bit 0 is the
// MSB; bit N sits in byte N/8 at position 7-(N%8).
func TestEncodeKDCOptionsBitLayout(t *testing.T) {
	cases := []struct {
		name string
		bits []int
		want []byte
	}{
		{"empty", nil, []byte{0x00, 0x00, 0x00, 0x00}},
		{"forwardable only (bit 1)", []int{kdcOptionForwardable}, []byte{0x40, 0x00, 0x00, 0x00}},
		{"proxiable only (bit 3)", []int{kdcOptionProxiable}, []byte{0x10, 0x00, 0x00, 0x00}},
		{"renewable only (bit 8)", []int{kdcOptionRenewable}, []byte{0x00, 0x80, 0x00, 0x00}},
		{"canonicalize only (bit 15)", []int{kdcOptionCanonicalize}, []byte{0x00, 0x01, 0x00, 0x00}},
		{"renewable-ok only (bit 27)", []int{kdcOptionRenewableOK}, []byte{0x00, 0x00, 0x00, 0x10}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := encodeKDCOptions(tc.bits...)
			if got.BitLength != 32 {
				t.Errorf("BitLength: got %d, want 32", got.BitLength)
			}
			if !bytes.Equal(got.Bytes, tc.want) {
				t.Errorf("Bytes: got %x, want %x", got.Bytes, tc.want)
			}
		})
	}
}

// TestKDCOptionsForASReq verifies the exact wire-format bits an AS-REQ
// should carry: forwardable + proxiable + renewable.
func TestKDCOptionsForASReq(t *testing.T) {
	// forwardable (bit 1 → byte 0 0x40) + proxiable (bit 3 → byte 0 0x10)
	// + renewable (bit 8 → byte 1 0x80)
	want := []byte{0x50, 0x80, 0x00, 0x00}
	got := kdcOptionsForASReq()
	if got.BitLength != 32 {
		t.Errorf("BitLength: got %d, want 32", got.BitLength)
	}
	if !bytes.Equal(got.Bytes, want) {
		t.Errorf("Bytes: got %x, want %x", got.Bytes, want)
	}
}

// TestKDCOptionsForTGSReq verifies the exact wire-format bits a TGS-REQ
// should carry: forwardable + renewable + canonicalize + renewable-ok.
func TestKDCOptionsForTGSReq(t *testing.T) {
	// forwardable (bit 1 → byte 0 0x40) + renewable (bit 8 → byte 1 0x80)
	// + canonicalize (bit 15 → byte 1 0x01) + renewable-ok (bit 27 → byte 3 0x10)
	want := []byte{0x40, 0x81, 0x00, 0x10}
	got := kdcOptionsForTGSReq()
	if got.BitLength != 32 {
		t.Errorf("BitLength: got %d, want 32", got.BitLength)
	}
	if !bytes.Equal(got.Bytes, want) {
		t.Errorf("Bytes: got %x, want %x", got.Bytes, want)
	}
}
