package rpcinterface_3919286ab10c11d09ba800c04fd92ef5_0_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the dssetup interface
// (3919286a-b10c-11d0-9ba8-00c04fd92ef5 v0.0, [MS-DSSP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "3919286a-b10c-11d0-9ba8-00c04fd92ef5" {
		t.Errorf("UUID = %s, want 3919286a-b10c-11d0-9ba8-00c04fd92ef5", got)
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the transport endpoint. dssetup shares \lsarpc ([MS-DSSP] 2.1).
func TestPipeName(t *testing.T) {
	if PipeName != `\lsarpc` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\lsarpc`)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// the single on-the-wire opnum (0) is covered; opnums 1..11 are NotUsedOnWire and absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 1 {
		t.Fatalf("OpnumToName has %d entries, want 1 (only opnum 0 is on the wire)", len(OpnumToName))
	}
	if OpnumToName[OpnumDsRolerGetPrimaryDomainInformation] != "DsRolerGetPrimaryDomainInformation" {
		t.Errorf("opnum 0 = %q, want DsRolerGetPrimaryDomainInformation", OpnumToName[OpnumDsRolerGetPrimaryDomainInformation])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestStatusString checks a known mnemonic and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(ErrorAccessDenied); got != "ERROR_ACCESS_DENIED" {
		t.Errorf("StatusString(5) = %q, want ERROR_ACCESS_DENIED", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}
