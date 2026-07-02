package rpcinterface_50abc2a4574d40b39d66ee4fd5fba076_5_0

import "testing"

// TestSyntaxID pins the abstract syntax: 50abc2a4-574d-40b3-9d66-ee4fd5fba076 v5.0.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	u := s.UUID
	if u.A != 0x50abc2a4 || u.B != 0x574d || u.C != 0x40b3 || u.D != 0x9d66 || u.E != 0xee4fd5fba076 {
		t.Errorf("UUID = %s, want 50abc2a4-574d-40b3-9d66-ee4fd5fba076", u.ToFormatD())
	}
	if s.MajorVersion != 5 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 5.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent and that all
// 19 on-the-wire opnums (0..18) are present exactly once.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 19 {
		t.Fatalf("OpnumToName has %d entries, want 19", len(OpnumToName))
	}
	seen := make(map[uint16]bool)
	for op, name := range OpnumToName {
		if op > 18 {
			t.Errorf("opnum %d out of range 0..18", op)
		}
		seen[op] = true
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	for op := uint16(0); op <= 18; op++ {
		if !seen[op] {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusString spot-checks a few mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:            "ERROR_SUCCESS",
		DnsErrorZoneDoesNotExist: "DNS_ERROR_ZONE_DOES_NOT_EXIST",
		DnsErrorDpAlreadyExists:  "DNS_ERROR_DP_ALREADY_EXISTS",
		0xdeadbeef:               "0xdeadbeef",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
}
