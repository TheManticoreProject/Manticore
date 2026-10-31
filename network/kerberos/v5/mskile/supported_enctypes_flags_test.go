package mskile

import "testing"

// TestSupportedEncryptionTypesFlagValues pins every flag to the value given by
// the MS-KILE 2.2.7 bit diagram (bits numbered MSB-first: diagram bit n is
// 1 << (31-n)).
func TestSupportedEncryptionTypesFlagValues(t *testing.T) {
	diagramBit := func(n uint) int { return 1 << (31 - n) }
	for _, tc := range []struct {
		name string
		got  int
		want int
	}{
		{"A DES-CBC-CRC", EncTypeDESCBCCRC, diagramBit(31)},
		{"B DES-CBC-MD5", EncTypeDESCBCMD5, diagramBit(30)},
		{"C RC4-HMAC", EncTypeRC4HMAC, diagramBit(29)},
		{"D AES128-CTS-HMAC-SHA1-96", EncTypeAES128CTSHMACSHA196, diagramBit(28)},
		{"E AES256-CTS-HMAC-SHA1-96", EncTypeAES256CTSHMACSHA196, diagramBit(27)},
		{"F FAST-supported", EncTypeFASTSupported, diagramBit(15)},
		{"G Compound-identity-supported", EncTypeCompoundIdentity, diagramBit(14)},
		{"H Claims-supported", EncTypeClaimsSupported, diagramBit(13)},
		{"I Resource-SID-compression-disabled", EncTypeResourceSIDCompDisabled, diagramBit(12)},
		{"J AES256-CTS-HMAC-SHA1-96-SK", EncTypeAES256SK, diagramBit(26)},
		{"K AES128-CTS-HMAC-SHA256-128", EncTypeAES128CTSHMACSHA256, diagramBit(25)},
		{"L AES256-CTS-HMAC-SHA384-192", EncTypeAES256CTSHMACSHA384, diagramBit(24)},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = 0x%08x, want 0x%08x", tc.name, tc.got, tc.want)
		}
	}
}
