package mskile

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

func TestPACRequestMatchesKnownEncoding(t *testing.T) {
	// The canonical PA-PAC-REQUEST(TRUE): SEQUENCE { [0] BOOLEAN TRUE }.
	want := []byte{0x30, 0x05, 0xa0, 0x03, 0x01, 0x01, 0xff}
	got, err := PACRequest(true)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("PACRequest(true) = % X, want % X", got, want)
	}

	inc, err := ParsePACRequest(got)
	if err != nil || !inc {
		t.Errorf("ParsePACRequest round-trip: inc=%v err=%v", inc, err)
	}

	falseBytes, _ := PACRequest(false)
	if inc, _ := ParsePACRequest(falseBytes); inc {
		t.Error("PACRequest(false) parsed as true")
	}

	pa, err := PACRequestPAData(true)
	if err != nil {
		t.Fatal(err)
	}
	if pa.PADataType != iana.PAPACRequest {
		t.Errorf("PA-DATA type = %d, want %d", pa.PADataType, iana.PAPACRequest)
	}
}

func TestPACOptionsRoundtrip(t *testing.T) {
	b, err := PACOptions(PACOptionResourceBasedConstrainedDeleg)
	if err != nil {
		t.Fatal(err)
	}
	// SEQUENCE { [0] BIT STRING(32) } — expect 30 07 a0 05 03 05 00 <4 bytes>.
	if b[0] != 0x30 || b[2] != 0xa0 || b[4] != 0x03 {
		t.Errorf("unexpected PA-PAC-OPTIONS framing: % X", b)
	}
	set, err := ParsePACOptions(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(set) != 1 || set[0] != PACOptionResourceBasedConstrainedDeleg {
		t.Errorf("parsed option bits = %v, want [%d]", set, PACOptionResourceBasedConstrainedDeleg)
	}

	multi, _ := PACOptions(PACOptionClaims, PACOptionForwardToFullDC)
	set2, _ := ParsePACOptions(multi)
	if len(set2) != 2 || set2[0] != PACOptionClaims || set2[1] != PACOptionForwardToFullDC {
		t.Errorf("multi-bit parse = %v", set2)
	}

	pa, err := PACOptionsPAData(PACOptionClaims)
	if err != nil {
		t.Fatal(err)
	}
	if pa.PADataType != iana.PAPACOptions {
		t.Errorf("PA-DATA type = %d, want %d", pa.PADataType, iana.PAPACOptions)
	}
}

func TestSupportedEnctypesRoundtrip(t *testing.T) {
	// DC functional level >= 2012 advertises 0x1F (DES..AES256).
	flags := int32(EncTypeRC4HMAC | EncTypeAES128CTSHMACSHA196 | EncTypeAES256CTSHMACSHA196)
	b, err := SupportedEnctypes(flags)
	if err != nil {
		t.Fatal(err)
	}
	// DER INTEGER: 0x1C -> 02 01 1C.
	if b[0] != 0x02 {
		t.Errorf("PA-SUPPORTED-ENCTYPES is not a DER INTEGER: % X", b)
	}
	got, err := ParseSupportedEnctypes(b)
	if err != nil {
		t.Fatal(err)
	}
	if got != flags {
		t.Errorf("round-trip flags = 0x%X, want 0x%X", got, flags)
	}

	// A larger value spanning multiple bytes (0x1F | claims | RBCD-ish).
	big := int32(0x1F | EncTypeClaimsSupported | EncTypeResourceSIDCompDisabled)
	bb, _ := SupportedEnctypes(big)
	if v, _ := ParseSupportedEnctypes(bb); v != big {
		t.Errorf("multi-byte round-trip = 0x%X, want 0x%X", v, big)
	}

	pa, err := SupportedEnctypesPAData(flags)
	if err != nil {
		t.Fatal(err)
	}
	if pa.PADataType != iana.PASupportedEnctypes {
		t.Errorf("PA-DATA type = %d, want %d", pa.PADataType, iana.PASupportedEnctypes)
	}
}
