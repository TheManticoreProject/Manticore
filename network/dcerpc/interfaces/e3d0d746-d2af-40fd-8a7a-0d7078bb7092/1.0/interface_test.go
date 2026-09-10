package rpcinterface_e3d0d746d2af40fd8a7a0d7078bb7092_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "e3d0d746-d2af-40fd-8a7a-0d7078bb7092" {
		t.Fatalf("UUID = %s, want e3d0d746-d2af-40fd-8a7a-0d7078bb7092", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the single opnum and its name mapping round-trips.
func TestOpnums(t *testing.T) {
	if OpnumExchangePublicKeys != 0 {
		t.Fatalf("OpnumExchangePublicKeys = %d, want 0", OpnumExchangePublicKeys)
	}
	if OpnumToName[0] != "ExchangePublicKeys" || NameToOpnum["ExchangePublicKeys"] != 0 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("map sizes differ: %d vs %d", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusCodesResolveThroughHRESULT pins that the two statuses this interface no
// longer declares resolve through the shared [MS-ERREF] 2.1.1 table, that StatusString
// defers to it, and that a value the specification does not define still renders as hex.
func TestStatusCodesResolveThroughHRESULT(t *testing.T) {
	// Zero is S_OK in the HRESULT space, the name the shared table renders for it; the
	// removed switch called it ERROR_SUCCESS, which is the Win32 name [MS-BPAU] 3.2.4.1
	// borrows in its Return Values list for the same value.
	if got := StatusString(0x00000000); got != "S_OK" {
		t.Errorf("StatusString(0x00000000) = %q, want S_OK", got)
	}

	// 0x80070005, the value [MS-BPAU] 3.2.4.1 gives for a caller whose Kerberos identity
	// is not in a trusted domain or whose certificate subject SID does not match it.
	if got := hresult.HRESULT(0x80070005).String(); got != "E_ACCESSDENIED" {
		t.Errorf("hresult.HRESULT(0x80070005).String() = %q, want E_ACCESSDENIED", got)
	}
	if got := StatusString(0x80070005); got != "E_ACCESSDENIED" {
		t.Errorf("StatusString(0x80070005) = %q, want E_ACCESSDENIED", got)
	}
	if resolved, defined := hresult.FromName("E_ACCESSDENIED"); !defined || resolved != 0x80070005 {
		t.Errorf("hresult.FromName(\"E_ACCESSDENIED\") = 0x%08x, %v; want 0x80070005, true", uint32(resolved), defined)
	}

	// E_FAIL is outside the three-value subset this interface used to declare, so it
	// rendered as hex; StatusString names it now.
	if got := StatusString(0x80004005); got != "E_FAIL" {
		t.Errorf("StatusString(0x80004005) = %q, want E_FAIL", got)
	}

	// A value [MS-ERREF] 2.1.1 does not define, and whose facility is not FACILITY_WIN32
	// either, still renders as hex.
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(0xdeadbeef) = %q, want hex", got)
	}
}

// TestStatusStringDecodesTableFullLocally pins the one reason StatusString still exists.
// 0x80040006 is in FACILITY_ITF, the interface-specific facility [MS-ERREF] 2.1 leaves to
// whichever interface returns the value, and the two claims on it disagree completely:
// [MS-BPAU] 3.2.4.1 returns it for a full peer-certificate table and gives it no symbolic
// name, while the shared table names it OLE_E_NOCACHE, "There is no cache to operate on".
// Routing this status through the table would report a certificate-table limit as an OLE
// caching complaint.
func TestStatusStringDecodesTableFullLocally(t *testing.T) {
	if StatusTableFull != 0x80040006 {
		t.Fatalf("StatusTableFull = 0x%08x, want 0x80040006", StatusTableFull)
	}
	if got := StatusString(StatusTableFull); got != "0x80040006 (peer-certificate table full)" {
		t.Errorf("StatusString(StatusTableFull) = %q, want the [MS-BPAU] meaning", got)
	}

	// The collision that makes the local constant necessary.
	entry, defined := hresult.Lookup(hresult.HRESULT(StatusTableFull))
	if !defined || entry.Name != "OLE_E_NOCACHE" {
		t.Errorf("hresult.Lookup(0x80040006) = %q, %v; want OLE_E_NOCACHE, true", entry.Name, defined)
	}
	if got := hresult.HRESULT(StatusTableFull).String(); got != "OLE_E_NOCACHE" {
		t.Errorf("hresult.HRESULT(0x80040006).String() = %q, want OLE_E_NOCACHE", got)
	}

	// The facility is 4, FACILITY_ITF, which is why both claims are legitimate.
	if facility := (StatusTableFull >> 16) & 0x07FF; facility != 4 {
		t.Errorf("facility of 0x80040006 = %d, want 4 (FACILITY_ITF)", facility)
	}
}
