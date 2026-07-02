package rpcinterface_91ae60209e3c11cf8d7c00aa00c091be_0_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(StatusInvalidArg); got != "E_INVALIDARG" {
		t.Errorf("StatusString(0x80070057) = %q, want E_INVALIDARG", got)
	}
	if got := StatusString(StatusAccessDenied); got != "E_ACCESSDENIED" {
		t.Errorf("StatusString(0x80000009) = %q, want E_ACCESSDENIED", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(unknown) = %q, want hex fallback", got)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	// ICertPassage defines a single on-the-wire method, CertServerRequest (opnum 0).
	if len(OpnumToName) != 1 {
		t.Errorf("OpnumToName has %d entries, want 1", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	if OpnumToName[OpnumCertServerRequest] != "CertServerRequest" || OpnumCertServerRequest != 0 {
		t.Errorf("OpnumToName[0] = %q (opnum %d), want CertServerRequest at 0",
			OpnumToName[OpnumCertServerRequest], OpnumCertServerRequest)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	// 91ae6020-9e3c-11cf-8d7c-00aa00c091be, version 0.0.
	if id.UUID.A != 0x91ae6020 || id.UUID.B != 0x9e3c || id.UUID.C != 0x11cf ||
		id.UUID.D != 0x8d7c || id.UUID.E != 0x00aa00c091be {
		t.Errorf("SyntaxID UUID = %+v, want 91ae6020-9e3c-11cf-8d7c-00aa00c091be", id.UUID)
	}
	if id.MajorVersion != 0 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 0.0", id.MajorVersion, id.MinorVersion)
	}
}
