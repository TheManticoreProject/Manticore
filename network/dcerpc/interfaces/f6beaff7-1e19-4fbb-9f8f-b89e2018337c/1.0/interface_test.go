package rpcinterface_f6beaff71e194fbb9f8fb89e2018337c_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax: f6beaff7-1e19-4fbb-9f8f-b89e2018337c v1.0.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	u := s.UUID
	if u.A != 0xf6beaff7 || u.B != 0x1e19 || u.C != 0x4fbb || u.D != 0x9f8f || u.E != 0xb89e2018337c {
		t.Errorf("UUID = %s, want f6beaff7-1e19-4fbb-9f8f-b89e2018337c", u.ToFormatD())
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the transport endpoint (RPC endpoint name "Eventlog", [MS-EVEN6]
// Standards Assignments).
func TestPipeName(t *testing.T) {
	if PipeName != `\eventlog` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\eventlog`)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent and that all
// 29 on-the-wire opnums (0..28, contiguous) are present exactly once.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 29 {
		t.Fatalf("OpnumToName has %d entries, want 29", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, want %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	for op := uint16(0); op <= 28; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("on-the-wire opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes [MS-EVEN6] 3.1.4
// documents for IEventService — the general [MS-ERREF] 2.2 codes and the ERROR_EVT_*
// range of the Windows Event Log — resolve through the shared table under their
// specification names, that a code the interface never enumerated now renders by name
// rather than as undecoded hex, and that a value the specification does not define still
// renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000003: "ERROR_PATH_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x00000103: "ERROR_NO_MORE_ITEMS",
		0x00000490: "ERROR_NOT_FOUND",
		0x000005B4: "ERROR_TIMEOUT",
		0x00003A98: "ERROR_EVT_INVALID_CHANNEL_PATH",
		0x00003A99: "ERROR_EVT_INVALID_QUERY",
		0x00003A9A: "ERROR_EVT_PUBLISHER_METADATA_NOT_FOUND",
		0x00003A9B: "ERROR_EVT_EVENT_TEMPLATE_NOT_FOUND",
		0x00003A9C: "ERROR_EVT_INVALID_PUBLISHER_NAME",
		0x00003A9D: "ERROR_EVT_INVALID_EVENT_DATA",
		0x00003A9F: "ERROR_EVT_CHANNEL_NOT_FOUND",
		0x00003AA0: "ERROR_EVT_MALFORMED_XML_TEXT",
		0x00003AA3: "ERROR_EVT_QUERY_RESULT_STALE",
		0x00003AB3: "ERROR_EVT_MESSAGE_NOT_FOUND",
		0x00003AB4: "ERROR_EVT_MESSAGE_ID_NOT_FOUND",
		0x00003AB5: "ERROR_EVT_UNRESOLVED_VALUE_INSERT",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// ERROR_EVT_CONFIGURATION_ERROR was outside the subset this interface used to
	// declare, so it rendered as hex; it resolves by name now.
	if got := win32.WIN32_ERROR(0x00003AA2).String(); got != "ERROR_EVT_CONFIGURATION_ERROR" {
		t.Errorf("win32.WIN32_ERROR(0x00003aa2).String() = %q, want ERROR_EVT_CONFIGURATION_ERROR", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
}
