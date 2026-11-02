package keycredentiallink

import (
	"encoding/hex"
	"io"
	"os"
	"testing"
)

// TestCheckIntegrityIsQuiet checks a failing CheckIntegrity returns false
// without writing to stdout: it used to print both hashes, which corrupts the
// output of any tool that calls it.
func TestCheckIntegrityIsQuiet(t *testing.T) {
	raw, err := hex.DecodeString(dsInternalsAzureADUserKey3)
	if err != nil {
		t.Fatal(err)
	}
	kc := &KeyCredentialLink{}
	if _, err := kc.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	kc.KeyHash = append([]byte(nil), kc.KeyHash...)
	kc.KeyHash[0] ^= 0xFF

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	ok := kc.CheckIntegrity()
	os.Stdout = stdout
	w.Close()
	printed, _ := io.ReadAll(r)

	if ok {
		t.Error("CheckIntegrity accepted a tampered KeyHash")
	}
	if len(printed) != 0 {
		t.Errorf("CheckIntegrity wrote to stdout: %q", printed)
	}
}
