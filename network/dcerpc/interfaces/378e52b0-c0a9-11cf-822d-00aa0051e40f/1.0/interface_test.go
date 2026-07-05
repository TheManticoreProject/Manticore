package rpcinterface_378e52b0c0a911cf822d00aa0051e40f_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of sasec.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "378e52b0-c0a9-11cf-822d-00aa0051e40f" {
		t.Fatalf("UUID = %s, want 378e52b0-c0a9-11cf-822d-00aa0051e40f", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName verifies the well-known ncacn_np endpoint shared with ATSvc.
func TestPipeName(t *testing.T) {
	if PipeName != `\atsvc` {
		t.Fatalf("PipeName = %q, want \\atsvc", PipeName)
	}
}

// TestOpnums verifies the implemented opnums and the name mapping round trip.
func TestOpnums(t *testing.T) {
	if OpnumSASetAccountInformation != 0 || OpnumSASetNSAccountInformation != 1 ||
		OpnumSAGetNSAccountInformation != 2 || OpnumSAGetAccountInformation != 3 {
		t.Fatalf("opnums = %d/%d/%d/%d, want 0/1/2/3",
			OpnumSASetAccountInformation, OpnumSASetNSAccountInformation,
			OpnumSAGetNSAccountInformation, OpnumSAGetAccountInformation)
	}
	if OpnumToName[0] != "SASetAccountInformation" || NameToOpnum["SAGetAccountInformation"] != 3 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "S_OK" {
		t.Fatalf("StatusString(StatusSuccess) = %s, want S_OK", got)
	}
	if got := StatusString(ErrorLogonFailure); got != "ERROR_LOGON_FAILURE" {
		t.Fatalf("StatusString(ErrorLogonFailure) = %s, want ERROR_LOGON_FAILURE", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}

// TestIsSuccess verifies the HRESULT success predicate (high bit clear).
func TestIsSuccess(t *testing.T) {
	if !IsSuccess(StatusSuccess) {
		t.Fatal("IsSuccess(S_OK) = false, want true")
	}
	if IsSuccess(ErrorAccessDenied) {
		t.Fatal("IsSuccess(E_ACCESSDENIED) = true, want false")
	}
}
