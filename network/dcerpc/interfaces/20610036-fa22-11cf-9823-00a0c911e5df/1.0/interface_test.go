package rpcinterface_20610036fa2211cf982300a0c911e5df_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the rasrpc interface
// (20610036-fa22-11cf-9823-00a0c911e5df v1.0, [MS-RRASM]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "20610036-fa22-11cf-9823-00a0c911e5df" {
		t.Errorf("UUID = %s, want 20610036-fa22-11cf-9823-00a0c911e5df", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// cover exactly the 7 on-the-wire opnums (5, 9, 10, 11, 12, 14, 15). Opnums 0-4, 6-8,
// 13, 16 are "not used on the wire" and are intentionally absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	wire := []uint16{5, 9, 10, 11, 12, 14, 15}
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
	for _, op := range []uint16{0, 1, 2, 3, 4, 6, 7, 8, 13, 16} {
		if _, ok := OpnumToName[op]; ok {
			t.Errorf("opnum %d is not used on the wire and must be absent", op)
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the eight Win32 codes this interface
// used to declare itself resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that a code the interface never enumerated now renders by
// name rather than as undecoded hex, and that a value the specification does not
// define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
	}
	for code, name := range documented {
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

	// ERROR_INVALID_LEVEL was outside the subset this interface used to declare, so
	// it rendered as hex; it resolves by name now, through StatusString as well.
	if got := win32.WIN32_ERROR(0x0000007C).String(); got != "ERROR_INVALID_LEVEL" {
		t.Errorf("win32.WIN32_ERROR(0x0000007c).String() = %q, want ERROR_INVALID_LEVEL", got)
	}
	if got := StatusString(0x0000007C); got != "ERROR_INVALID_LEVEL" {
		t.Errorf("StatusString(0x0000007c) = %q, want ERROR_INVALID_LEVEL", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(0xdeadbeef) = %q, want hex", got)
	}
}

// TestStatusStringDecodesRASSpecificError pins the one reason StatusString still
// exists: Remote Access errors live in the 600..999 range raserror.h owns, which
// [MS-ERREF] 2.2 does not cover, so routing StatusBufferTooSmall through the shared
// table would rename a RAS buffer-size complaint into an unrelated marshaling
// overflow.
func TestStatusStringDecodesRASSpecificError(t *testing.T) {
	if StatusBufferTooSmall != 0x0000025B {
		t.Fatalf("StatusBufferTooSmall = 0x%08x, want 0x0000025b (603, RASBASE+3)", StatusBufferTooSmall)
	}
	if StatusBufferTooSmall < 600 || StatusBufferTooSmall > 999 {
		t.Errorf("StatusBufferTooSmall = %d, outside the RAS error range 600..999", StatusBufferTooSmall)
	}
	if got := StatusString(StatusBufferTooSmall); got != "ERROR_BUFFER_TOO_SMALL" {
		t.Errorf("StatusString(0x0000025b) = %q, want ERROR_BUFFER_TOO_SMALL", got)
	}

	// The collision that makes the local constant necessary: the shared table names
	// 0x0000025B for the user/kernel marshaling buffer, not for RAS, and it knows no
	// code by the RAS name at all.
	if got := win32.WIN32_ERROR(StatusBufferTooSmall).String(); got != "ERROR_MARSHALL_OVERFLOW" {
		t.Errorf("win32.WIN32_ERROR(0x0000025b).String() = %q, want ERROR_MARSHALL_OVERFLOW", got)
	}
	if resolved, defined := win32.FromName("ERROR_BUFFER_TOO_SMALL"); defined {
		t.Errorf("win32.FromName(\"ERROR_BUFFER_TOO_SMALL\") = 0x%08x, true; want undefined", uint32(resolved))
	}
}

// TestPipeName pins the [MS-RRASM] 2.1 well-known endpoint shared with dimsvc.
func TestPipeName(t *testing.T) {
	if PipeName != `\ROUTER` {
		t.Errorf("PipeName = %q, want \\ROUTER", PipeName)
	}
}
