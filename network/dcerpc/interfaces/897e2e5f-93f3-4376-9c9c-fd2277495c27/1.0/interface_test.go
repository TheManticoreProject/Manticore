package rpcinterface_897e2e5f93f343769c9cfd2277495c27_1_0

import (
	"fmt"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the FrsTransport interface
// (897e2e5f-93f3-4376-9c9c-fd2277495c27 v1.0, [MS-FRS2]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "897e2e5f-93f3-4376-9c9c-fd2277495c27" {
		t.Errorf("UUID = %s, want 897e2e5f-93f3-4376-9c9c-fd2277495c27", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// every on-the-wire opnum is covered (opnum 14 is NotUsedOnWire and must be absent).
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 17 {
		t.Fatalf("OpnumToName has %d entries, want 17 on-the-wire opnums", len(OpnumToName))
	}
	if _, ok := OpnumToName[14]; ok {
		t.Errorf("opnum 14 is NotUsedOnWire but present in OpnumToName")
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that success, the one code this descriptor
// declared that [MS-ERREF] 2.2 defines, resolves through the shared table under its
// specification name, that generic Win32 codes the FrsTransport methods can return now
// render by name rather than as undecoded hex, and that a value the specification does
// not define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	if got := win32.WIN32_ERROR(0x00000000).String(); got != "ERROR_SUCCESS" {
		t.Errorf("win32.WIN32_ERROR(0x00000000).String() = %q, want ERROR_SUCCESS", got)
	}
	if resolved, defined := win32.FromName("ERROR_SUCCESS"); !defined || uint32(resolved) != 0x00000000 {
		t.Errorf("win32.FromName(\"ERROR_SUCCESS\") = 0x%08x, %v; want 0x00000000, true", uint32(resolved), defined)
	}
	if got := StatusString(0x00000000); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0x00000000) = %q, want ERROR_SUCCESS", got)
	}

	// Generic Win32 codes sit outside the three FRS-specific values the descriptor
	// declares, so StatusString used to render them as bare hex. They resolve by name now.
	generic := map[uint32]string{
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
	}
	for code, name := range generic {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(0xdeadbeef) = %q, want 0xdeadbeef", got)
	}
}

// TestStatusStringDecodesFrsSpecificErrors pins the reason StatusString still exists:
// [MS-ERREF] 2.2 has no row for any of the three FRS-specific values, so the shared
// table cannot name them, and the enclosing 0x2300..0x26FF block it does populate
// belongs to DNS, so the values must not be folded onto a neighbour.
func TestStatusStringDecodesFrsSpecificErrors(t *testing.T) {
	frsSpecific := map[uint32]string{
		FRS_ERROR_CONNECTION_INVALID:   "FRS_ERROR_CONNECTION_INVALID",
		FRS_ERROR_INCOMPATIBLE_VERSION: "FRS_ERROR_INCOMPATIBLE_VERSION",
		FRS_ERROR_CONTENTSET_READ_ONLY: "FRS_ERROR_CONTENTSET_READ_ONLY",
	}
	if len(frsSpecific) != 3 {
		t.Fatalf("FRS-specific table has %d entries, want 3", len(frsSpecific))
	}
	for code, name := range frsSpecific {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
		// The shared table claims neither the value nor the name, so nothing here
		// silently renames an FRS error into an unrelated one.
		if entry, defined := win32.Lookup(win32.WIN32_ERROR(code)); defined {
			t.Errorf("win32.Lookup(0x%08x) resolves to %q; %s must stay local", code, entry.Name, name)
		}
		if want := fmt.Sprintf("0x%08x", code); win32.WIN32_ERROR(code).String() != want {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, win32.WIN32_ERROR(code).String(), want)
		}
		if resolved, defined := win32.FromName(name); defined {
			t.Errorf("win32.FromName(%q) = 0x%08x; the code would belong in the shared table instead", name, uint32(resolved))
		}
	}

	// What [MS-ERREF] 2.2 does put in this range: DNS. Folding an FRS value onto a
	// populated neighbour would rename a replication error into a DNS complaint.
	if got := win32.WIN32_ERROR(0x0000232D).String(); got != "DNS_ERROR_RCODE_REFUSED" {
		t.Errorf("win32.WIN32_ERROR(0x0000232d).String() = %q, want DNS_ERROR_RCODE_REFUSED", got)
	}
	if got := win32.WIN32_ERROR(0x0000233A).String(); got != "DNS_ERROR_RCODE_BADTIME" {
		t.Errorf("win32.WIN32_ERROR(0x0000233a).String() = %q, want DNS_ERROR_RCODE_BADTIME", got)
	}
}
