package mslsad

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

func mustSID(t *testing.T, s string) *msdtyp.RPC_SID {
	t.Helper()
	sid, err := msdtyp.ParseSID(s)
	if err != nil {
		t.Fatalf("ParseSID(%q): %v", s, err)
	}
	return &sid
}

// TestLSAPR_CR_CIPHER_VALUE_RoundTrip exercises a conformant-varying counted byte blob.
func TestLSAPR_CR_CIPHER_VALUE_RoundTrip(t *testing.T) {
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02}
	in := LSAPR_CR_CIPHER_VALUE{
		Length:        uint32(len(payload)),
		MaximumLength: uint32(len(payload)),
		Buffer:        payload,
	}
	roundTrip(t, "LSAPR_CR_CIPHER_VALUE", in)
}

// TestLSAPR_PRIVILEGE_SET_RoundTrip exercises an inline conformant array (not a pointer)
// whose maximum_count is hoisted to the front of the structure.
func TestLSAPR_PRIVILEGE_SET_RoundTrip(t *testing.T) {
	in := LSAPR_PRIVILEGE_SET{
		PrivilegeCount: 2,
		Control:        1,
		Privilege: []LSAPR_LUID_AND_ATTRIBUTES{
			{Luid: msdtyp.LUID{LowPart: 0x14, HighPart: 0}, Attributes: 3},
			{Luid: msdtyp.LUID{LowPart: 0x11, HighPart: 0}, Attributes: 0},
		},
	}
	roundTrip(t, "LSAPR_PRIVILEGE_SET", in)
}

// TestLSAPR_AES_CIPHER_VALUE_RoundTrip exercises fixed 64/16-byte arrays followed by a
// [size_is(cbCipher)] [unique] pointer to a conformant byte array.
func TestLSAPR_AES_CIPHER_VALUE_RoundTrip(t *testing.T) {
	var auth [64]uint8
	var salt [16]uint8
	for i := range auth {
		auth[i] = uint8(i)
	}
	for i := range salt {
		salt[i] = uint8(0x80 + i)
	}
	in := LSAPR_AES_CIPHER_VALUE{
		AuthData: auth,
		Salt:     salt,
		CbCipher: 4,
		Cipher:   []uint8{0xde, 0xad, 0xbe, 0xef},
	}
	roundTrip(t, "LSAPR_AES_CIPHER_VALUE", in)
}

// TestLSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL_AES_RoundTrip exercises the AES
// auth-blob variant, which shares LSAPR_AES_CIPHER_VALUE's layout.
func TestLSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL_AES_RoundTrip(t *testing.T) {
	in := LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL_AES{
		CbCipher: 3,
		Cipher:   []uint8{1, 2, 3},
	}
	roundTrip(t, "LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL_AES", in)
}

// TestLSAPR_REVISION_INFO_RoundTrip exercises the negotiation union with its V1 arm
// selected; the ULONG discriminant is transmitted inline ahead of the arm.
func TestLSAPR_REVISION_INFO_RoundTrip(t *testing.T) {
	in := LSAPR_REVISION_INFO{
		Tag: 1,
		V1:  LSAPR_REVISION_INFO_V1{Revision: 1, SupportedFeatures: 0x00000001},
	}
	roundTrip(t, "LSAPR_REVISION_INFO", in)
}

// TestLSA_FOREST_TRUST_SCANNER_INFO_RoundTrip exercises a [unique] RPC_SID pointer plus
// two counted Unicode strings.
func TestLSA_FOREST_TRUST_SCANNER_INFO_RoundTrip(t *testing.T) {
	in := LSA_FOREST_TRUST_SCANNER_INFO{
		DomainSid:   mustSID(t, "S-1-5-21-7-8-9"),
		DnsName:     msdtyp.NewUnicodeString("contoso.com"),
		NetbiosName: msdtyp.NewUnicodeString("CONTOSO"),
	}
	roundTrip(t, "LSA_FOREST_TRUST_SCANNER_INFO", in)
}

// TestLSA_FOREST_TRUST_INFORMATION2_RoundTrip exercises a [unique] pointer to a conformant
// array of [unique] pointers to LSA_FOREST_TRUST_RECORD2, selecting a different union arm
// (DomainInfo=2, ScannerInfo=4, TopLevelName=0) in each record. Only the field matching a
// record's ForestTrustType is populated so the round trip is deeply equal.
func TestLSA_FOREST_TRUST_INFORMATION2_RoundTrip(t *testing.T) {
	in := LSA_FOREST_TRUST_INFORMATION2{
		RecordCount: 3,
		Entries: []*LSA_FOREST_TRUST_RECORD2{
			{
				Flags:           0x1,
				ForestTrustType: ForestTrustDomainInfo,
				Time:            msdtyp.LARGE_INTEGER(0x0000000512340000),
				ForestTrustData: LSA_FOREST_TRUST_DATA2{
					ForestTrustType: ForestTrustDomainInfo,
					DomainInfo: LSA_FOREST_TRUST_DOMAIN_INFO{
						Sid:         mustSID(t, "S-1-5-21-1-2-3"),
						DnsName:     msdtyp.NewUnicodeString("child.contoso.com"),
						NetbiosName: msdtyp.NewUnicodeString("CHILD"),
					},
				},
			},
			{
				Flags:           0x0,
				ForestTrustType: ForestTrustScannerInfo,
				ForestTrustData: LSA_FOREST_TRUST_DATA2{
					ForestTrustType: ForestTrustScannerInfo,
					ScannerInfo: LSA_FOREST_TRUST_SCANNER_INFO{
						DomainSid:   mustSID(t, "S-1-5-21-4-5-6"),
						DnsName:     msdtyp.NewUnicodeString("fabrikam.com"),
						NetbiosName: msdtyp.NewUnicodeString("FABRIKAM"),
					},
				},
			},
			{
				ForestTrustType: ForestTrustTopLevelName,
				ForestTrustData: LSA_FOREST_TRUST_DATA2{
					ForestTrustType: ForestTrustTopLevelName,
					TopLevelName:    msdtyp.NewUnicodeString("contoso.com"),
				},
			},
		},
	}
	roundTrip(t, "LSA_FOREST_TRUST_INFORMATION2", in)
}
