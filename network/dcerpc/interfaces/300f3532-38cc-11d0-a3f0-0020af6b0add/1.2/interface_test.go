package rpcinterface_300f353238cc11d0a3f00020af6b0add_1_2

import "testing"

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

// TestStatusString checks the documented mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:          "S_OK",
		TrkEReferral:           "TRK_E_REFERRAL",
		TrkEPotentialFileFound: "TRK_E_POTENTIAL_FILE_FOUND",
		EptSNotRegistered:      "EPT_S_NOT_REGISTERED",
		0xdeadbeef:             "0xdeadbeef",
	}
	for status, want := range cases {
		if got := StatusString(status); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", status, got, want)
		}
	}
}

// TestStatusIsSuccess pins the HRESULT sign-bit success rule ([MS-DLTW] 3.1.4.1): zero and
// positive values succeed; the TRK_E_* soft failures (sign bit set) do not.
func TestStatusIsSuccess(t *testing.T) {
	if !StatusIsSuccess(StatusSuccess) {
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
}
