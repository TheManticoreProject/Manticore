package rpcinterface_367abb81984435f1ad3298f038001003_2_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes [MS-SCMR] 3.1.4
// documents for svcctl resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that a code the interface never enumerated now renders by name
// rather than as undecoded hex, and that a value the specification does not define
// still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000003: "ERROR_PATH_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x0000000D: "ERROR_INVALID_DATA",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x0000007B: "ERROR_INVALID_NAME",
		0x0000007C: "ERROR_INVALID_LEVEL",
		0x000000EA: "ERROR_MORE_DATA",
		0x0000041B: "ERROR_DEPENDENT_SERVICES_RUNNING",
		0x0000041C: "ERROR_INVALID_SERVICE_CONTROL",
		0x0000041D: "ERROR_SERVICE_REQUEST_TIMEOUT",
		0x0000041E: "ERROR_SERVICE_NO_THREAD",
		0x0000041F: "ERROR_SERVICE_DATABASE_LOCKED",
		0x00000420: "ERROR_SERVICE_ALREADY_RUNNING",
		0x00000421: "ERROR_INVALID_SERVICE_ACCOUNT",
		0x00000422: "ERROR_SERVICE_DISABLED",
		0x00000423: "ERROR_CIRCULAR_DEPENDENCY",
		0x00000424: "ERROR_SERVICE_DOES_NOT_EXIST",
		0x00000425: "ERROR_SERVICE_CANNOT_ACCEPT_CTRL",
		0x00000426: "ERROR_SERVICE_NOT_ACTIVE",
		0x00000429: "ERROR_DATABASE_DOES_NOT_EXIST",
		0x0000042C: "ERROR_SERVICE_DEPENDENCY_FAIL",
		0x0000042D: "ERROR_SERVICE_LOGON_FAILED",
		0x00000430: "ERROR_SERVICE_MARKED_FOR_DELETE",
		0x00000431: "ERROR_SERVICE_EXISTS",
		0x00000433: "ERROR_SERVICE_DEPENDENCY_DELETED",
		0x00000436: "ERROR_DUPLICATE_SERVICE_NAME",
		0x0000045B: "ERROR_SHUTDOWN_IN_PROGRESS",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// ERROR_SERVICE_START_HANG was outside the subset this interface used to declare, so
	// it rendered as hex; it resolves by name now.
	if got := win32.WIN32_ERROR(0x0000042E).String(); got != "ERROR_SERVICE_START_HANG" {
		t.Errorf("win32.WIN32_ERROR(0x0000042e).String() = %q, want ERROR_SERVICE_START_HANG", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if id.UUID.A != 0x367abb81 || id.UUID.B != 0x9844 || id.UUID.C != 0x35f1 ||
		id.UUID.D != 0xad32 || id.UUID.E != 0x98f038001003 {
		t.Errorf("SyntaxID UUID = %+v, want 367abb81-9844-35f1-ad32-98f038001003", id.UUID)
	}
	if id.MajorVersion != 2 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 2.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	if len(OpnumToName) != 50 {
		t.Errorf("OpnumToName has %d entries, want 50 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	// Spot-check a few opnum numbers against [MS-SCMR] 3.1.4.
	if OpnumToName[OpnumROpenSCManagerW] != "ROpenSCManagerW" || OpnumROpenSCManagerW != 15 {
		t.Errorf("OpnumROpenSCManagerW = %d (%q)", OpnumROpenSCManagerW, OpnumToName[OpnumROpenSCManagerW])
	}
	if OpnumToName[OpnumRCloseServiceHandle] != "RCloseServiceHandle" || OpnumRCloseServiceHandle != 0 {
		t.Errorf("OpnumRCloseServiceHandle = %d", OpnumRCloseServiceHandle)
	}
	if OpnumROpenSCManager2 != 64 {
		t.Errorf("OpnumROpenSCManager2 = %d, want 64", OpnumROpenSCManager2)
	}
}

// TestNotUsedOnWireOmitted asserts the 15 "not used on the wire" opnums are absent.
func TestNotUsedOnWireOmitted(t *testing.T) {
	for _, op := range []uint16{10, 22, 34, 43, 46, 52, 53, 54, 55, 57, 58, 59, 61, 62, 63} {
		if name, ok := OpnumToName[op]; ok {
			t.Errorf("opnum %d should be NotUsedOnWire but maps to %q", op, name)
		}
	}
}
