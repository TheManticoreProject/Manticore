package customkeyinformation

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

// [MS-ADTS] 2.2.20.2 long form: Version(1) Flags(1) VolumeType(1)
// SupportsNotification(1) FekKeyVersion(1) KeyStrength(4) Reserved(10)
// EncodedExtendedCKI(variable). The Reserved bytes must be consumed before the
// extended information is read.
func TestCustomKeyInformationParsesReserved(t *testing.T) {
	kcv := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}
	data := []byte{1, 0, 0, 1, 1, 2, 0, 0, 0}
	data = append(data, 0x05, 0, 0, 0, 0, 0, 0, 0, 0, 0x07) // Reserved
	data = append(data, 0, 2, 0xAA, 0xBB)                   // EncodedExtendedCKI: version 0, size 2

	var cki CustomKeyInformation
	n, err := cki.Unmarshal(data, kcv)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != len(data) {
		t.Errorf("Unmarshal consumed %d bytes, want %d", n, len(data))
	}
	if cki.Reserved == nil || cki.Reserved[0] != 0x05 || cki.Reserved[9] != 0x07 {
		t.Errorf("Reserved = %v, want the 10 bytes after KeyStrength", cki.Reserved)
	}
	if cki.EncodedExtendedCKI == nil || !bytes.Equal(cki.EncodedExtendedCKI.Data, []byte{0xAA, 0xBB}) {
		t.Errorf("EncodedExtendedCKI = %+v, want Data aabb", cki.EncodedExtendedCKI)
	}
	out, err := cki.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(out, data) {
		t.Errorf("round trip mismatch:\n got %x\nwant %x", out, data)
	}
}
