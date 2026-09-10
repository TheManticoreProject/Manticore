package rpcinterface_a8e0653c27444389a61d7373df8b2292_1_0

import (
	"fmt"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// TestSyntaxID confirms the abstract syntax is the FileServerVssAgent UUID/version
// (a8e0653c-2744-4389-a61d-7373df8b2292 v1.0, [MS-FSRVP] 2.1).
func TestSyntaxID(t *testing.T) {
	want, err := guid.FromString("a8e0653c-2744-4389-a61d-7373df8b2292")
	if err != nil {
		t.Fatalf("FromString: %v", err)
	}
	id := SyntaxID()
	if id.UUID != *want {
		t.Errorf("UUID = %s, want %s", id.UUID.ToFormatD(), want.ToFormatD())
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName covers all 13 on-the-wire methods and that
// NameToOpnum is a faithful inverse.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 13 {
		t.Fatalf("OpnumToName has %d entries, want 13", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, want %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d (ok=%v), want %d", name, got, ok, op)
		}
	}
}

// TestMigratedStatusCodesResolveThroughHRESULT pins the three values this descriptor no
// longer declares. Each has a row in [MS-ERREF] 2.1.1 under the name FSRVP uses for it,
// so it resolves through the shared table in both directions and through StatusString's
// fallthrough. It also pins that a common HRESULT outside the old subset now renders by
// name, and that an undefined value still renders as hex.
func TestMigratedStatusCodesResolveThroughHRESULT(t *testing.T) {
	migrated := map[hresult.HRESULT]string{
		hresult.E_INVALIDARG:   "E_INVALIDARG",
		hresult.E_ACCESSDENIED: "E_ACCESSDENIED",
		hresult.E_OUTOFMEMORY:  "E_OUTOFMEMORY",
	}
	values := map[string]uint32{
		"E_INVALIDARG":   0x80070057,
		"E_ACCESSDENIED": 0x80070005,
		"E_OUTOFMEMORY":  0x8007000E,
	}
	for code, name := range migrated {
		if got := uint32(code); got != values[name] {
			t.Errorf("hresult.%s = 0x%08x, want 0x%08x", name, got, values[name])
		}
		if got := code.String(); got != name {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want %q", uint32(code), got, name)
		}
		if resolved, defined := hresult.FromName(name); !defined || resolved != code {
			t.Errorf("hresult.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, uint32(code))
		}
		if got := StatusString(uint32(code)); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", uint32(code), got, name)
		}
	}

	// E_UNEXPECTED was outside the subset this descriptor used to declare, so it
	// rendered as hex; it resolves by name now, through StatusString as well.
	if got := StatusString(0x8000FFFF); got != "E_UNEXPECTED" {
		t.Errorf("StatusString(0x8000ffff) = %q, want E_UNEXPECTED", got)
	}
	// A value [MS-ERREF] 2.1.1 does not define still renders as hex.
	if got := StatusString(0xdeadbeef); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
	// Zero keeps the specification's own name for it, not the shared table's S_OK.
	if got := StatusString(StatusSuccess); got != "ZERO" {
		t.Errorf("StatusString(0) = %q, want ZERO", got)
	}
	if got := hresult.HRESULT(StatusSuccess).String(); got != "S_OK" {
		t.Errorf("hresult.HRESULT(0).String() = %q, want S_OK", got)
	}
}

// TestStatusStringDecodesFsrvpSpecificErrors pins the reason StatusString still exists.
// Six of these codes sit in FACILITY_ITF, the interface-specific facility, where
// [MS-FSRVP] defines its own meanings and [MS-ERREF] 2.1.1 has no row at all; a seventh
// sits at 0x80042501 in the same facility; and the two wait results are Win32 values
// carried verbatim, one of them success-severity. None can be read out of the shared
// table, so all nine must keep decoding here.
func TestStatusStringDecodesFsrvpSpecificErrors(t *testing.T) {
	fsrvpSpecific := map[uint32]string{
		FsrvpEBadState:                "FSRVP_E_BAD_STATE",
		FsrvpENotSupported:            "FSRVP_E_NOT_SUPPORTED",
		FsrvpEObjectAlreadyExists:     "FSRVP_E_OBJECT_ALREADY_EXISTS",
		FsrvpEObjectNotFound:          "FSRVP_E_OBJECT_NOT_FOUND",
		FsrvpEShadowCopySetInProgress: "FSRVP_E_SHADOW_COPY_SET_IN_PROGRESS",
		FsrvpEUnsupportedContext:      "FSRVP_E_UNSUPPORTED_CONTEXT",
		FsrvpEShadowcopysetIdMismatch: "FSRVP_E_SHADOWCOPYSET_ID_MISMATCH",
		FsrvpEWaitTimeout:             "FSRVP_E_WAIT_TIMEOUT",
		FsrvpEWaitFailed:              "FSRVP_E_WAIT_FAILED",
	}
	if len(fsrvpSpecific) != 9 {
		t.Fatalf("FSRVP-specific table has %d entries, want 9", len(fsrvpSpecific))
	}
	for code, name := range fsrvpSpecific {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
		if _, defined := hresult.Lookup(hresult.HRESULT(code)); defined {
			t.Errorf("hresult.Lookup(0x%08x) resolves; %s belongs in the shared table instead", code, name)
		}
		if _, defined := hresult.FromName(name); defined {
			t.Errorf("hresult.FromName(%q) resolves; the code belongs in the shared table instead", name)
		}
		if want := fmt.Sprintf("0x%08x", code); hresult.HRESULT(code).String() != want {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want the hex rendering %q", code, hresult.HRESULT(code).String(), want)
		}
	}

	// The seven HRESULT-shaped codes are all in FACILITY_ITF (4), the facility whose
	// meanings are the interface's own to define; that is why the table cannot name them.
	for _, code := range []uint32{
		FsrvpEBadState, FsrvpENotSupported, FsrvpEObjectAlreadyExists, FsrvpEObjectNotFound,
		FsrvpEShadowCopySetInProgress, FsrvpEUnsupportedContext, FsrvpEShadowcopysetIdMismatch,
	} {
		if facility := (code >> 16) & 0x07FF; facility != 0x004 {
			t.Errorf("0x%08x is in facility %d, want FACILITY_ITF (4)", code, facility)
		}
		if hresult.HRESULT(code).IsSuccess() {
			t.Errorf("HRESULT(0x%08x).IsSuccess() = true, want false", code)
		}
	}

	// FSRVP_E_WAIT_TIMEOUT is the value that makes the stubs' comparison against
	// StatusSuccess load-bearing: its severity bit is clear, so the shared table's
	// IsSuccess calls a timed-out wait a success.
	if !hresult.HRESULT(FsrvpEWaitTimeout).IsSuccess() {
		t.Errorf("HRESULT(0x%08x).IsSuccess() = false, want true", FsrvpEWaitTimeout)
	}
	if err := hresult.HRESULT(FsrvpEWaitTimeout).Error(); err != nil {
		t.Errorf("HRESULT(0x%08x).Error() = %v, want nil for a success-severity value", FsrvpEWaitTimeout, err)
	}
}
