package customkeyinformation

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/strength"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

// TestCustomKeyInformationKeepsDataAfterReserved checks every byte after a full
// Reserved field survives Unmarshal and Marshal: a lone trailing byte used to be
// dropped, and an EncodedExtendedCKI with an unknown version or a Size past the
// end made Unmarshal fail, which rejected the whole credential.
func TestCustomKeyInformationKeepsDataAfterReserved(t *testing.T) {
	kcv := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}
	head := "010000010102" + "00000000000000000000" // through a full Reserved
	for _, tc := range []struct {
		name, tail string
		wantExt    bool
		wantTrail  string
	}{
		{"one trailing byte", "ee", false, "ee"},
		{"extension version 1", "0102aabb", false, "0102aabb"},
		{"extension size past the end", "0005aabb", false, "0005aabb"},
		{"valid extension", "0002aabb", true, ""},
		{"valid extension then bytes", "0002aabbccdd", true, "ccdd"},
	} {
		data, _ := hex.DecodeString(head + tc.tail)
		var cki CustomKeyInformation
		n, err := cki.Unmarshal(data, kcv)
		if err != nil {
			t.Fatalf("%s: Unmarshal: %v", tc.name, err)
		}
		if n != len(data) {
			t.Errorf("%s: consumed %d of %d bytes", tc.name, n, len(data))
		}
		if (cki.EncodedExtendedCKI != nil) != tc.wantExt {
			t.Errorf("%s: EncodedExtendedCKI = %+v", tc.name, cki.EncodedExtendedCKI)
		}
		if got := hex.EncodeToString(cki.Trailing); got != tc.wantTrail {
			t.Errorf("%s: Trailing = %q, want %q", tc.name, got, tc.wantTrail)
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

// TestCustomKeyInformationTrailingNeedsFullReserved checks Marshal refuses trailing
// bytes after a short Reserved, which a reader would take as part of Reserved.
func TestCustomKeyInformationTrailingNeedsFullReserved(t *testing.T) {
	short := Reserved{1, 2}
	cki := CustomKeyInformation{Version: 1, Reserved: &short, Trailing: []byte{0xee}}
	flags := CUSTOMKEYINFO_FLAGS(0)
	cki.Flags = &flags
	vt := CustomKeyInformationVolumeType(0)
	cki.VolumeType = &vt
	sn := SupportsNotification(false)
	cki.SupportsNotification = &sn
	fv := FekKeyVersion(1)
	cki.FekKeyVersion = &fv
	ks := strength.KeyStrength(0)
	cki.KeyStrength = &ks
	if out, err := cki.Marshal(); err == nil {
		t.Errorf("Marshal wrote %x, want an error for trailing data after a 2-byte Reserved", out)
	}
}

// TestCustomKeyInformationUnmarshalClearsTrailing checks a reused value does not
// keep the Trailing bytes of an earlier parse.
func TestCustomKeyInformationUnmarshalClearsTrailing(t *testing.T) {
	kcv := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}
	withTail, _ := hex.DecodeString("010000010102" + "00000000000000000000" + "ee")
	without, _ := hex.DecodeString("010000010102" + "00000000000000000000")
	var cki CustomKeyInformation
	if _, err := cki.Unmarshal(withTail, kcv); err != nil {
		t.Fatal(err)
	}
	if _, err := cki.Unmarshal(without, kcv); err != nil {
		t.Fatal(err)
	}
	if cki.Trailing != nil {
		t.Errorf("Trailing = %x after parsing a value without trailing bytes", cki.Trailing)
	}
	if out, _ := cki.Marshal(); !bytes.Equal(out, without) {
		t.Errorf("Marshal = %x, want %x", out, without)
	}
}
