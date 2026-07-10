package pkinit

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

// randomBytes returns n cryptographically random bytes.
func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("pkinit: read random: %w", err)
	}
	return b, nil
}

// GenerateSelfSignedCert generates a fresh RSA key pair and a self-signed X.509
// certificate over it, suitable for the Shadow Credentials technique: the
// certificate's public key is what gets registered in a target's
// msDS-KeyCredentialLink, and the private key signs the PKINIT AuthPack. The
// certificate is not validated by the KDC in the key-trust model (the KDC maps
// the client via the registered key, not a PKI chain), so a self-signed
// certificate with an arbitrary subject is sufficient.
//
// bits is the RSA modulus size (2048 recommended). subjectCN is the certificate
// subject common name (cosmetic).
func GenerateSelfSignedCert(bits int, subjectCN string) (*rsa.PrivateKey, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("pkinit: generate RSA key: %w", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("pkinit: generate serial: %w", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: subjectCN},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("pkinit: create self-signed certificate: %w", err)
	}
	return priv, certDER, nil
}
