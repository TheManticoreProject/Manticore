package attacks

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// Exact example hashes from the hashcat wiki (example_hashes), used as vectors.
const (
	exASREP23 = "$krb5asrep$23$user@domain.com:3e156ada591263b8aab0965f5aebd837$007497cb51b6c8116d6407a782ea0e1c5402b17db7afa6b05a6d30ed164a9933c754d720e279c6c573679bd27128fe77e5fea1f72334c1193c8ff0b370fadc6368bf2d49bbfdba4c5dccab95e8c8ebfdc75f438a0797dbfb2f8a1a5f4c423f9bfc1fea483342a11bd56a216f4d5158ccc4b224b52894fadfba3957dfe4b6b8f5f9f9fe422811a314768673e0c924340b8ccb84775ce9defaa3baa0910b676ad0036d13032b0dd94e3b13903cc738a7b6d00b0b3c210d1f972a6c7cae9bd3c959acf7565be528fc179118f28c679f6deeee1456f0781eb8154e18e49cb27b64bf74cd7112a0ebae2102ac"
	exTGS23   = "$krb5tgs$23$*user$realm$test/spn*$63386d22d359fe42230300d56852c9eb$891ad31d09ab89c6b3b8c5e5de6c06a7f49fd559d7a9a3c32576c8fedf705376cea582ab5938f7fc8bc741acf05c5990741b36ef4311fe3562a41b70a4ec6ecba849905f2385bb3799d92499909658c7287c49160276bca0006c350b0db4fd387adc27c01e9e9ad0c20ed53a7e6356dee2452e35eca2a6a1d1432796fc5c19d068978df74d3d0baf35c77de12456bf1144b6a750d11f55805f5a16ece2975246e2d026dce997fba34ac8757312e9e4e6272de35e20d52fb668c5ed"
	exTGS17   = "$krb5tgs$17$user$realm$ae8434177efd09be5bc2eff8$90b4ce5b266821adc26c64f71958a475cf9348fce65096190be04f8430c4e0d554c86dd7ad29c275f9e8f15d2dab4565a3d6e21e449dc2f88e52ea0402c7170ba74f4af037c5d7f8db6d53018a564ab590fc23aa1134788bcc4a55f69ec13c0a083291a96b41bffb978f5a160b7edc828382d11aacd89b5a1bfa710b0e591b190bff9062eace4d26187777db358e70efd26df9c9312dbeef20b1ee0d823d4e71b8f1d00d91ea017459c27c32dc20e451ea6278be63cdd512ce656357c942b95438228e"
	exTGS18   = "$krb5tgs$18$user$realm$8efd91bb01cc69dd07e46009$7352410d6aafd72c64972a66058b02aa1c28ac580ba41137d5a170467f06f17faf5dfb3f95ecf4fad74821fdc7e63a3195573f45f962f86942cb24255e544ad8d05178d560f683a3f59ce94e82c8e724a3af0160be549b472dd83e6b80733ad349973885e9082617294c6cbbea92349671883eaf068d7f5dcfc0405d97fda27435082b82b24f3be27f06c19354bf32066933312c770424eb6143674756243c1bde78ee3294792dcc49008a1b54f32ec5d5695f899946d42a67ce2fb1c227cb1d2004c0"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex: %v", err)
	}
	return b
}

// rebuildRC4 reconstructs the KDC-REP ciphertext for an RC4 hash: checksum(16) || edata.
func rebuildRC4(t *testing.T, cksumHex, edataHex string) []byte {
	return append(mustHex(t, cksumHex), mustHex(t, edataHex)...)
}

// rebuildAES reconstructs the ciphertext for an AES hash: edata || checksum.
func rebuildAES(t *testing.T, cksumHex, edataHex string) []byte {
	return append(mustHex(t, edataHex), mustHex(t, cksumHex)...)
}

func TestFormatASREPHashRC4(t *testing.T) {
	// $krb5asrep$23$user@domain.com:<cksum>$<edata>
	rest := exASREP23[strings.Index(exASREP23, ":")+1:]
	cksumHex, edataHex, _ := strings.Cut(rest, "$")
	cipher := rebuildRC4(t, cksumHex, edataHex)

	got, err := FormatASREPHash("user", "domain.com", iana.ETypeRC4HMAC, cipher)
	if err != nil {
		t.Fatal(err)
	}
	if got != exASREP23 {
		t.Errorf("AS-REP RC4 hash mismatch:\n got  %s\n want %s", got, exASREP23)
	}
}

// TestFormatASREPHashRejectsAES verifies that FormatASREPHash refuses every AES
// enctype: hashcat has no AS-REP mode for AES (mode 18200 is RC4/etype-23 only),
// so an "$krb5asrep$17/18/19/20$" string would be uncrackable. AES AS-REP hashes
// must go through FormatASREPHashJohn instead.
func TestFormatASREPHashRejectsAES(t *testing.T) {
	for _, tc := range []struct {
		name  string
		etype int
	}{
		{"aes128-sha1", iana.ETypeAES128CTSHMACSHA196},
		{"aes256-sha1", iana.ETypeAES256CTSHMACSHA196},
		{"aes256-sha2", iana.ETypeAES256CTSHMACSHA384},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := FormatASREPHash("user", "domain.com", tc.etype, make([]byte, 64)); err == nil {
				t.Errorf("expected error for AES etype %d, got nil", tc.etype)
			}
		})
	}
}

func lastTwoFields(s string) (secondLast, last string) {
	f := strings.Split(s, "$")
	return f[len(f)-2], f[len(f)-1]
}

func TestFormatTGSHashRC4(t *testing.T) {
	cksumHex, edataHex := lastTwoFields(exTGS23)
	cipher := rebuildRC4(t, cksumHex, edataHex)

	got, err := FormatTGSHash("user", "realm", "test/spn", iana.ETypeRC4HMAC, cipher)
	if err != nil {
		t.Fatal(err)
	}
	if got != exTGS23 {
		t.Errorf("TGS RC4 hash mismatch:\n got  %s\n want %s", got, exTGS23)
	}
}

func TestFormatTGSHashAES(t *testing.T) {
	for _, tc := range []struct {
		name  string
		etype int
		ex    string
	}{
		{"aes128", iana.ETypeAES128CTSHMACSHA196, exTGS17},
		{"aes256", iana.ETypeAES256CTSHMACSHA196, exTGS18},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cksumHex, edataHex := lastTwoFields(tc.ex)
			cipher := rebuildAES(t, cksumHex, edataHex)
			got, err := FormatTGSHash("user", "realm", "ignored/spn", tc.etype, cipher)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.ex {
				t.Errorf("TGS %s hash mismatch:\n got  %s\n want %s", tc.name, got, tc.ex)
			}
		})
	}
}

func TestFormatRejectsShortCipher(t *testing.T) {
	if _, err := FormatASREPHash("u", "R", iana.ETypeRC4HMAC, []byte{1, 2, 3}); err == nil {
		t.Error("expected error for short RC4 cipher")
	}
	if _, err := FormatTGSHash("u", "R", "s", iana.ETypeAES256CTSHMACSHA196, []byte{1, 2, 3}); err == nil {
		t.Error("expected error for short AES cipher")
	}
	if _, err := FormatTGSHash("u", "R", "s", 999, make([]byte, 32)); err == nil {
		t.Error("expected error for unsupported etype")
	}
}
