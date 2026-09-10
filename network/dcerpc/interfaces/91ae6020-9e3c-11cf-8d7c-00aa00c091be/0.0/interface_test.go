package rpcinterface_91ae60209e3c11cf8d7c00aa00c091be_0_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
)

// TestStatusCodesResolveThroughHRESULT pins that the two statuses this interface no longer
// declares resolve through the shared [MS-ERREF] 2.1.1 table, that StatusString defers to
// it, and that a value the specification does not define still renders as hex.
func TestStatusCodesResolveThroughHRESULT(t *testing.T) {
	// Zero is S_OK in the HRESULT space; [MS-ICPR] 3.2.4.1 calls the same value
	// ERROR_SUCCESS, which is the Win32 name the removed switch rendered.
	if got := StatusString(0x00000000); got != "S_OK" {
		t.Errorf("StatusString(0x00000000) = %q, want S_OK", got)
	}

	// E_INVALIDARG (0x80070057) is the failure [MS-ICPR] 3.2.4.1 documents when the cb
	// field of pctbAttribs does not match the length of the string it points to. Value and
	// name agree with the shared table in both directions.
	if got := hresult.HRESULT(0x80070057).String(); got != "E_INVALIDARG" {
		t.Errorf("hresult.HRESULT(0x80070057).String() = %q, want E_INVALIDARG", got)
	}
	if got := StatusString(0x80070057); got != "E_INVALIDARG" {
		t.Errorf("StatusString(0x80070057) = %q, want E_INVALIDARG", got)
	}
	if resolved, defined := hresult.FromName("E_INVALIDARG"); !defined || resolved != 0x80070057 {
		t.Errorf("hresult.FromName(\"E_INVALIDARG\") = 0x%08x, %v; want 0x80070057, true", uint32(resolved), defined)
	}

	// The enrollment failures ICertPassage inherits from ICertRequestD::Request are
	// outside the three-value subset this interface used to declare, so they rendered as
	// hex; StatusString names them now.
	for code, name := range map[uint32]string{
		0x80094800: "CERTSRV_E_UNSUPPORTED_CERT_TYPE",
		0x80094012: "CERTSRV_E_TEMPLATE_DENIED",
	} {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
	}

	// A value [MS-ERREF] 2.1.1 does not define still renders as hex, as it did before.
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(0x12345678) = %q, want hex fallback", got)
	}
}

// TestStatusStringDecodesLegacyAccessDeniedLocally pins the one reason StatusString still
// exists. [MS-ICPR] 3.2.4.1 says the CA "MUST refuse to establish a connection with the
// client by returning E_ACCESSDENIED (0x80000009)" when packet privacy is required and was
// not negotiated, but that pairing is the legacy OLE E_ACCESSDENIED of the 0x8000000X
// block: [MS-ERREF] 2.1.1 defines no 0x80000009 at all and gives the name E_ACCESSDENIED
// to 0x80070005 instead. The value has to stay verbatim to match what a CA returns, and
// the shared table cannot name it.
func TestStatusStringDecodesLegacyAccessDeniedLocally(t *testing.T) {
	if StatusAccessDenied != 0x80000009 {
		t.Fatalf("StatusAccessDenied = 0x%08x, want 0x80000009 ([MS-ICPR] 3.2.4.1)", StatusAccessDenied)
	}
	if got := StatusString(StatusAccessDenied); got != "E_ACCESSDENIED" {
		t.Errorf("StatusString(0x80000009) = %q, want E_ACCESSDENIED", got)
	}

	// The disagreement that makes the local constant necessary: the shared table defines
	// nothing for this value and holds the name against a different one.
	if entry, defined := hresult.Lookup(hresult.HRESULT(StatusAccessDenied)); defined {
		t.Errorf("hresult.Lookup(0x80000009) = %q, true; want undefined", entry.Name)
	}
	if got := hresult.HRESULT(StatusAccessDenied).String(); got != "0x80000009" {
		t.Errorf("hresult.HRESULT(0x80000009).String() = %q, want hex", got)
	}
	if resolved, defined := hresult.FromName("E_ACCESSDENIED"); !defined || resolved != 0x80070005 {
		t.Errorf("hresult.FromName(\"E_ACCESSDENIED\") = 0x%08x, %v; want 0x80070005, true: the table gives the name to a different value", uint32(resolved), defined)
	}

	// It is a failure in either reading, so a caller that only tests the severity bit
	// still treats it as one.
	if hresult.HRESULT(StatusAccessDenied).IsSuccess() {
		t.Error("hresult.HRESULT(0x80000009).IsSuccess() = true, want false")
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
