package customkeyinformation

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

// TestCustomKeyInformationRealEntriesRoundTrip parses the CustomKeyInformation
// entries of two real Azure AD keys from DSInternals' KeyCredentialTester.cs and
// checks they marshal back unchanged. KeyStrength is one byte, and the FIDO2
// entry ends nine bytes into the 10-byte Reserved field; reading KeyStrength as
// four bytes and dropping a partial Reserved shortened both on marshal.
func TestCustomKeyInformationRealEntriesRoundTrip(t *testing.T) {
	kcv := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}
	for _, tc := range []struct {
		name, hexData string
		strength      uint32
		reservedLen   int
	}{
		{"UserKey2", "010000010000", 0, 0},
		{"UserKeyFIDO0", "010100000000000000000000000000", 0, 9},
	} {
		data, _ := hex.DecodeString(tc.hexData)
		var cki CustomKeyInformation
		n, err := cki.Unmarshal(data, kcv)
		if err != nil {
			t.Fatalf("%s: Unmarshal: %v", tc.name, err)
		}
		if n != len(data) {
			t.Errorf("%s: consumed %d of %d bytes", tc.name, n, len(data))
		}
		if cki.KeyStrength == nil || uint32(*cki.KeyStrength) != tc.strength {
			t.Errorf("%s: KeyStrength = %v", tc.name, cki.KeyStrength)
		}
		gotReserved := 0
		if cki.Reserved != nil {
			gotReserved = len(*cki.Reserved)
		}
		if gotReserved != tc.reservedLen {
			t.Errorf("%s: Reserved is %d bytes, want %d", tc.name, gotReserved, tc.reservedLen)
		}
		out, err := cki.Marshal()
		if err != nil {
			t.Fatalf("%s: Marshal: %v", tc.name, err)
		}
		if !bytes.Equal(out, data) {
			t.Errorf("%s: round trip\n got  %x\n want %x", tc.name, out, data)
		}
	}
}

// TestCustomKeyInformationMarshalRejectsShortReservedWithExtension checks Marshal
// refuses EncodedExtendedCKI after a Reserved field shorter than 10 bytes: a
// reader would take the extension bytes as the rest of Reserved.
func TestCustomKeyInformationMarshalRejectsShortReservedWithExtension(t *testing.T) {
	data, _ := hex.DecodeString("010000000000" + "00000000000000000000" + "0001aa")
	var cki CustomKeyInformation
	if _, err := cki.Unmarshal(data, version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if cki.EncodedExtendedCKI == nil {
		t.Fatal("no EncodedExtendedCKI parsed")
	}
	short := Reserved{1, 2, 3}
	cki.Reserved = &short
	if out, err := cki.Marshal(); err == nil {
		t.Errorf("Marshal wrote %x, want an error for a 3-byte Reserved before EncodedExtendedCKI", out)
	}
}
