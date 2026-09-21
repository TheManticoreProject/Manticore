package rpcinterface_82273fdce32a18c33f78827929dc23ea_0_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax: 82273fdc-e32a-18c3-3f78-827929dc23ea v0.0.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	u := s.UUID
	if u.A != 0x82273fdc || u.B != 0xe32a || u.C != 0x18c3 || u.D != 0x3f78 || u.E != 0x827929dc23ea {
		t.Errorf("UUID = %s, want 82273fdc-e32a-18c3-3f78-827929dc23ea", u.ToFormatD())
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the transport endpoint.
func TestPipeName(t *testing.T) {
	if PipeName != `\eventlog` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\eventlog`)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent, that the
// 23 on-the-wire opnums are present exactly once, and that the four "not used on the
// wire" opnums (19, 20, 21, 23) are absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	wireOpnums := []uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 22, 24, 25, 26}
	if len(OpnumToName) != len(wireOpnums) {
		t.Fatalf("OpnumToName has %d entries, want %d", len(OpnumToName), len(wireOpnums))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	for _, op := range wireOpnums {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("on-the-wire opnum %d missing from OpnumToName", op)
		}
	}
	for _, gap := range []uint16{19, 20, 21, 23} {
		if name, ok := OpnumToName[gap]; ok {
			t.Errorf("opnum %d (%q) is NotUsedOnWire and must be absent", gap, name)
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes [MS-EVEN] 3.1.4
// documents for the EventLog Remoting Protocol resolve through the shared [MS-ERREF]
// 2.2 table under their specification names, that a code this interface never
// enumerated now renders by name rather than as undecoded hex, and that a value the
// specification does not define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x00000026: "ERROR_HANDLE_EOF",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x000005DC: "ERROR_EVENTLOG_FILE_CORRUPT",
		0x000005DD: "ERROR_EVENTLOG_CANT_START",
		0x000005DE: "ERROR_LOG_FILE_FULL",
		0x000005DF: "ERROR_EVENTLOG_FILE_CHANGED",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// ERROR_MORE_DATA was outside the subset this interface used to declare, so it
	// rendered as hex; it resolves by name now.
	if got := win32.WIN32_ERROR(0x000000EA).String(); got != "ERROR_MORE_DATA" {
		t.Errorf("win32.WIN32_ERROR(0x000000ea).String() = %q, want ERROR_MORE_DATA", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xdeadbeef).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}

// TestStatusStringRoutesNTStatusThroughNTStatusTable verifies that the two NTSTATUS
// values this interface can return are routed through the nt_status table, Win32
// codes go through the win32 table, and unknown values render as hex.
func TestStatusStringRoutesNTStatusThroughNTStatusTable(t *testing.T) {
	// The two NTSTATUS values resolve through the nt_status table.
	if got := StatusString(0xC0000023); got != nt_status.NT_STATUS_BUFFER_TOO_SMALL.String() {
		t.Errorf("StatusString(0xc0000023) = %q, want %q", got, nt_status.NT_STATUS_BUFFER_TOO_SMALL.String())
	}
	if got := StatusString(0xC000000D); got != nt_status.NT_STATUS_INVALID_PARAMETER.String() {
		t.Errorf("StatusString(0xc000000d) = %q, want %q", got, nt_status.NT_STATUS_INVALID_PARAMETER.String())
	}

	// Win32 codes still resolve through the win32 table.
	if got := StatusString(0x00000000); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0x00000000) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(0x000005DF); got != "ERROR_EVENTLOG_FILE_CHANGED" {
		t.Errorf("StatusString(0x000005df) = %q, want ERROR_EVENTLOG_FILE_CHANGED", got)
	}
	if got := StatusString(0x0000000D); got != "ERROR_INVALID_DATA" {
		t.Errorf("StatusString(0x0000000d) = %q, want ERROR_INVALID_DATA", got)
	}

	// NTSTATUS with severity bit set but not in the table renders as hex via nt_status.
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(0xdeadbeef) = %q, want hex fallback", got)
	}
}
