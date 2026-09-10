package rpcinterface_300f353238cc11d0a3f00020af6b0add_1_2

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
)

// TestSyntaxID pins the abstract syntax identifier for the trkwks interface
// (300f3532-38cc-11d0-a3f0-0020af6b0add v1.2, [MS-DLTW]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "300f3532-38cc-11d0-a3f0-0020af6b0add" {
		t.Errorf("UUID = %s, want 300f3532-38cc-11d0-a3f0-0020af6b0add", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 2 {
		t.Errorf("version = %d.%d, want 1.2", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// the single on-the-wire opnum (12, LnkSearchMachine) is covered — opnums 0..11 are
// "not used on the wire" and are intentionally absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 1 {
		t.Fatalf("OpnumToName has %d entries, want 1 (opnum 12 only)", len(OpnumToName))
	}
	if OpnumToName[OpnumLnkSearchMachine] != "LnkSearchMachine" {
		t.Errorf("opnum 12 = %q, want LnkSearchMachine", OpnumToName[OpnumLnkSearchMachine])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestStatusCodesResolveThroughHRESULT pins that the success value this interface used to
// declare resolves through the shared [MS-ERREF] 2.1.1 table, that StatusString defers to
// it, and that a value the specification does not define still renders as hex.
func TestStatusCodesResolveThroughHRESULT(t *testing.T) {
	if got := hresult.S_OK.String(); got != "S_OK" {
		t.Errorf("hresult.S_OK.String() = %q, want S_OK", got)
	}
	if got := StatusString(0x00000000); got != "S_OK" {
		t.Errorf("StatusString(0x00000000) = %q, want S_OK", got)
	}
	if resolved, defined := hresult.FromName("S_OK"); !defined || resolved != 0x00000000 {
		t.Errorf("hresult.FromName(\"S_OK\") = 0x%08x, %v; want 0x00000000, true", uint32(resolved), defined)
	}

	// E_ACCESSDENIED and E_OUTOFMEMORY are outside the four-value subset this interface
	// used to declare, so they rendered as hex; StatusString names them now.
	for code, name := range map[uint32]string{0x80070005: "E_ACCESSDENIED", 0x8007000E: "E_OUTOFMEMORY"} {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
	}

	// A value [MS-ERREF] 2.1.1 does not define, and whose facility is not FACILITY_WIN32
	// either, still renders as hex.
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(0xdeadbeef) = %q, want hex", got)
	}
}

// TestStatusStringDecodesForeignSpacesLocally pins the three statuses StatusString still
// decodes itself. TRK_E_REFERRAL and TRK_E_POTENTIAL_FILE_FOUND come from a private
// Distributed Link Tracking facility that [MS-ERREF] 2.1.1 does not cover, and
// EPT_S_NOT_REGISTERED is the NTSTATUS-shaped RPC endpoint-mapper failure of [MS-ERREF]
// 2.3. The shared table names none of the three, so routing them through it would turn
// three meaningful statuses — two of which carry [out] parameters callers act on — back
// into hex.
func TestStatusStringDecodesForeignSpacesLocally(t *testing.T) {
	local := map[uint32]struct{ name, hex string }{
		TrkEReferral:           {"TRK_E_REFERRAL", "0x8dead101"},
		TrkEPotentialFileFound: {"TRK_E_POTENTIAL_FILE_FOUND", "0x8dead106"},
		EptSNotRegistered:      {"EPT_S_NOT_REGISTERED", "0xc0020017"},
	}
	for code, want := range local {
		if got := StatusString(code); got != want.name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want.name)
		}
		if entry, defined := hresult.Lookup(hresult.HRESULT(code)); defined {
			t.Errorf("hresult.Lookup(0x%08x) = %q, true; want undefined, so the local name must stand", code, entry.Name)
		}
		if got := hresult.HRESULT(code).String(); got != want.hex {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want %q", code, got, want.hex)
		}
		if resolved, defined := hresult.FromName(want.name); defined {
			t.Errorf("hresult.FromName(%q) = 0x%08x, true; want undefined", want.name, uint32(resolved))
		}
	}

	// The values themselves, and the facility field that keeps the two TRK_E_ statuses out
	// of the shared table: 0x5EA is not a facility [MS-ERREF] 2.1 assigns.
	if TrkEReferral != 0x8DEAD101 || TrkEPotentialFileFound != 0x8DEAD106 {
		t.Fatalf("TRK_E_ values = 0x%08x/0x%08x, want 0x8dead101/0x8dead106", TrkEReferral, TrkEPotentialFileFound)
	}
	for _, code := range []uint32{TrkEReferral, TrkEPotentialFileFound} {
		if facility := (code >> 16) & 0x07FF; facility != 0x5EA {
			t.Errorf("facility of 0x%08x = 0x%03x, want 0x5ea", code, facility)
		}
	}
	if EptSNotRegistered != 0xC0020017 {
		t.Errorf("EptSNotRegistered = 0x%08x, want 0xc0020017", EptSNotRegistered)
	}
}

// TestStatusIsSuccess pins the HRESULT sign-bit success rule ([MS-DLTW] 3.1.4.1): zero and
// any positive value are success, a sign-bit-set value is a failure.
func TestStatusIsSuccess(t *testing.T) {
	if !StatusIsSuccess(uint32(hresult.S_OK)) {
		t.Errorf("StatusIsSuccess(S_OK) = false, want true")
	}
	if !StatusIsSuccess(0x00000001) {
		t.Errorf("StatusIsSuccess(positive) = false, want true")
	}
	if StatusIsSuccess(TrkEReferral) {
		t.Errorf("StatusIsSuccess(TRK_E_REFERRAL) = true, want false")
	}
	if StatusIsSuccess(TrkEPotentialFileFound) {
		t.Errorf("StatusIsSuccess(TRK_E_POTENTIAL_FILE_FOUND) = true, want false")
	}
	if StatusIsSuccess(EptSNotRegistered) {
		t.Errorf("StatusIsSuccess(EPT_S_NOT_REGISTERED) = true, want false")
	}
}
