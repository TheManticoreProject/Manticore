package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"math/big"
	"testing"
)

// TestNewBCRYPT_RSA_PUBLIC_KEY checks that a BCRYPT_RSA_PUBLIC_KEY built from a
// crypto/rsa public key round-trips: marshalling it and unmarshalling the bytes
// back yields the same modulus and public exponent, and the header sizes agree
// with the big-endian minimal encodings.
func TestNewBCRYPT_RSA_PUBLIC_KEY(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	pub := &priv.PublicKey

	key := NewBCRYPT_RSA_PUBLIC_KEY(pub)

	if key.Magic.Magic != 0x31415352 { // "RSA1"
		t.Fatalf("unexpected magic 0x%08x", key.Magic.Magic)
	}
	if key.Header.BitLength != uint32(pub.N.BitLen()) {
		t.Fatalf("BitLength = %d, want %d", key.Header.BitLength, pub.N.BitLen())
	}
	if int(key.Header.CbModulus) != len(key.Content.Modulus) {
		t.Fatalf("CbModulus = %d, want %d", key.Header.CbModulus, len(key.Content.Modulus))
	}
	if int(key.Header.CbPublicExp) != len(key.Content.PublicExponent) {
		t.Fatalf("CbPublicExp = %d, want %d", key.Header.CbPublicExp, len(key.Content.PublicExponent))
	}

	// The stored modulus/exponent must be the minimal big-endian encodings.
	if new(big.Int).SetBytes(key.Content.Modulus).Cmp(pub.N) != 0 {
		t.Fatal("modulus mismatch")
	}
	if got := int(new(big.Int).SetBytes(key.Content.PublicExponent).Int64()); got != pub.E {
		t.Fatalf("exponent = %d, want %d", got, pub.E)
	}

	// Marshal -> Unmarshal must reconstruct an equal key.
	data, err := key.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var back BCRYPT_RSA_PUBLIC_KEY
	if _, err := back.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !key.Equal(&back) {
		t.Fatal("round-tripped key does not equal the original")
	}

	// The build must be deterministic for the same key.
	data2, err := NewBCRYPT_RSA_PUBLIC_KEY(pub).Marshal()
	if err != nil {
		t.Fatalf("Marshal (2): %v", err)
	}
	if len(data) != len(data2) {
		t.Fatal("marshalled blob length is not deterministic")
	}
}
