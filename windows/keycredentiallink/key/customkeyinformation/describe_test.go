package customkeyinformation

import (
	"encoding/hex"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

// TestDescribePrintsFieldValues checks Describe prints the numeric values of
// VolumeType, FekKeyVersion and KeyStrength. It used to hand the pointer fields
// to %02x / %d, printing the hex of each field's name ("0x4e6f6e65" for None)
// or a pointer.
func TestDescribePrintsFieldValues(t *testing.T) {
	// Version 1, Flags 0, VolumeType 1 (OS volume), SupportsNotification 1,
	// FekKeyVersion 1, KeyStrength 2 (Normal).
	data, _ := hex.DecodeString("010001010102")
	var cki CustomKeyInformation
	if _, err := cki.Unmarshal(data, version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	cki.Describe(0)
	os.Stdout = stdout
	w.Close()
	raw, _ := io.ReadAll(r)
	out := string(raw)

	for _, want := range []string{"(0x01)\n", "FekKeyVersion\x1b[0m: 1\n", "(0x02)\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("Describe output lacks %q:\n%s", want, out)
		}
	}
	for _, bad := range []string{"0x4f", "0x4e6f6e65", "0x4e6f726d616c"} {
		if strings.Contains(out, bad) {
			t.Errorf("Describe output contains %q (a field name printed as hex):\n%s", bad, out)
		}
	}
}
