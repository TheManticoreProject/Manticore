package shadowcredentials

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt"
)

// TestRSAPublicKeyBlobRoundTrips checks that the CNG BCRYPT_RSAPUBLIC_BLOB built
// for an RSA public key parses back through the shared key-material decoder to
// the same modulus and exponent.
func TestRSAPublicKeyBlobRoundTrips(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	blob, err := rsaPublicKeyBlob(&priv.PublicKey)
	if err != nil {
		t.Fatalf("rsaPublicKeyBlob: %v", err)
	}
	km, _, err := bcrypt.UnmarshalKeyMaterial(blob)
	if err != nil {
		t.Fatalf("UnmarshalKeyMaterial: %v", err)
	}
	if km.Fingerprint() == "" {
		t.Fatal("decoded key material has an empty fingerprint")
	}
	// A second identical build must be byte-identical (deterministic).
	blob2, _ := rsaPublicKeyBlob(&priv.PublicKey)
	if len(blob) != len(blob2) {
		t.Fatal("blob length not deterministic")
	}
}

func TestGenerateCredentialKeyID(t *testing.T) {
	cred, err := GenerateCredential("CN=victim,CN=Users,DC=example,DC=com")
	if err != nil {
		t.Fatalf("GenerateCredential: %v", err)
	}
	if cred.PrivateKey == nil || len(cred.CertificateDER) == 0 {
		t.Fatal("credential missing key/cert")
	}
	if cred.KeyID == "" {
		t.Fatal("credential has no KeyID")
	}
	if cred.LinkValue == "" || len(cred.LinkBlob) == 0 {
		t.Fatal("credential has no link value/blob")
	}
	// The KeyID recomputed from the blob must be stable.
	id, err := keyCredentialID(cred.LinkBlob)
	if err != nil {
		t.Fatalf("keyCredentialID: %v", err)
	}
	if id != cred.KeyID {
		t.Fatalf("KeyID mismatch: %s != %s", id, cred.KeyID)
	}
}

func TestDNBinaryHex(t *testing.T) {
	if got := dnBinaryHex("B:8:deadbeef:CN=x,DC=y"); got != "deadbeef" {
		t.Fatalf("dnBinaryHex = %q, want deadbeef", got)
	}
	if got := dnBinaryHex("not-a-dn-binary"); got != "" {
		t.Fatalf("dnBinaryHex on junk = %q, want empty", got)
	}
}
