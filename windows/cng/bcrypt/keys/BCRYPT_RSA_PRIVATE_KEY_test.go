package keys

import "testing"

// A blob whose prime is 1 makes p-1 or q-1 zero; ExportDER must reject it instead
// of dividing by zero while computing the CRT exponents.
func TestBCRYPT_RSA_PRIVATE_KEY_ExportDERRejectsPrimeOne(t *testing.T) {
	// RSA2, BitLength 8, cbPublicExp 1, cbModulus 1, cbPrime1 1, cbPrime2 1;
	// e = 1, n = 0x0f, p = 0x05, q = 0x01.
	blob := []byte("RSA2\x08\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x01\x0f\x05\x01")
	var key BCRYPT_RSA_PRIVATE_KEY
	if _, err := key.Unmarshal(blob); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, err := key.ExportDER(); err == nil {
		t.Fatal("ExportDER accepted a prime of 1")
	}
}
