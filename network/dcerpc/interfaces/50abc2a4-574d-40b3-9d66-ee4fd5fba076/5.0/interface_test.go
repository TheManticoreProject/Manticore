package rpcinterface_50abc2a4574d40b39d66ee4fd5fba076_5_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

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

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes [MS-DNSP] documents for
// the DnsServer methods resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that a DNS_ERROR_* code the interface never enumerated now
// renders by name rather than as undecoded hex, and that a value the specification does
// not define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x00002329: "DNS_ERROR_RCODE_FORMAT_ERROR",
		0x0000232A: "DNS_ERROR_RCODE_SERVER_FAILURE",
		0x0000232B: "DNS_ERROR_RCODE_NAME_ERROR",
		0x0000232C: "DNS_ERROR_RCODE_NOT_IMPLEMENTED",
		0x0000232D: "DNS_ERROR_RCODE_REFUSED",
		0x0000254F: "DNS_ERROR_INVALID_TYPE",
		0x00002550: "DNS_ERROR_INVALID_IP_ADDRESS",
		0x00002581: "DNS_ERROR_ZONE_DOES_NOT_EXIST",
		0x00002582: "DNS_ERROR_NO_ZONE_INFO",
		0x00002583: "DNS_ERROR_INVALID_ZONE_OPERATION",
		0x00002584: "DNS_ERROR_ZONE_CONFIGURATION_ERROR",
		0x00002589: "DNS_ERROR_ZONE_ALREADY_EXISTS",
		0x0000258B: "DNS_ERROR_INVALID_ZONE_TYPE",
		0x000025E5: "DNS_ERROR_RECORD_DOES_NOT_EXIST",
		0x000025EF: "DNS_ERROR_RECORD_ALREADY_EXISTS",
		0x000025F2: "DNS_ERROR_NAME_DOES_NOT_EXIST",
		0x000025F5: "DNS_ERROR_DS_UNAVAILABLE",
		0x000026AD: "DNS_ERROR_DP_DOES_NOT_EXIST",
		0x000026AE: "DNS_ERROR_DP_ALREADY_EXISTS",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// DNS_ERROR_ZONE_LOCKED was outside the subset this interface used to declare, so it
	// rendered as hex; it resolves by name now.
	if got := win32.WIN32_ERROR(0x00002587).String(); got != "DNS_ERROR_ZONE_LOCKED" {
		t.Errorf("win32.WIN32_ERROR(0x00002587).String() = %q, want DNS_ERROR_ZONE_LOCKED", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xdeadbeef).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}
