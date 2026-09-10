package rpcinterface_d95afe70a6d54259822e2c84da1ddb0d_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the WindowsShutdown interface
// (d95afe70-a6d5-4259-822e-2c84da1ddb0d v1.0, [MS-RSP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "d95afe70-a6d5-4259-822e-2c84da1ddb0d" {
		t.Errorf("UUID = %s, want d95afe70-a6d5-4259-822e-2c84da1ddb0d", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// cover exactly the 2 on-the-wire opnums (0, 1).
func TestOpnumNameRoundTrip(t *testing.T) {
	wire := []uint16{0, 1}
	if len(OpnumToName) != len(wire) {
		t.Fatalf("OpnumToName has %d entries, want %d", len(OpnumToName), len(wire))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for _, op := range wire {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes [MS-RSP] section 3.3.4
// documents for WindowsShutdown resolve through the shared [MS-ERREF] 2.2 table under
// their specification names, that a shutdown code the interface never enumerated now
// renders by name rather than as undecoded hex, and that a value the specification
// does not define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000015: "ERROR_NOT_READY",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000045B: "ERROR_SHUTDOWN_IN_PROGRESS",
		0x0000045C: "ERROR_NO_SHUTDOWN_IN_PROGRESS",
		0x000004A6: "ERROR_SHUTDOWN_IS_SCHEDULED",
		0x000004A7: "ERROR_SHUTDOWN_USERS_LOGGED_ON",
		0x000004F7: "ERROR_MACHINE_LOCKED",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// ERROR_SYSTEM_SHUTDOWN was outside the subset this interface used to declare, so a
	// server reporting it produced bare hex; it resolves by name now.
	if got := win32.WIN32_ERROR(0x00000281).String(); got != "ERROR_SYSTEM_SHUTDOWN" {
		t.Errorf("win32.WIN32_ERROR(0x00000281).String() = %q, want ERROR_SYSTEM_SHUTDOWN", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}

// TestPipeName pins the empty endpoint: WindowsShutdown is ncacn_ip_tcp / dynamic
// endpoint per [MS-RSP] 2.1, so it has no well-known named pipe.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (TCP dynamic endpoint)", PipeName)
	}
}
