package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// X509Certificate represents an X.509 certificate along with its associated RSA private key and public key material.
//
// Fields:
// - key: A pointer to an rsa.PrivateKey object representing the RSA private key associated with the certificate.
// - certificate: A pointer to an x509.Certificate object representing the X.509 certificate.
// - publicKey: An RSAKeyMaterial object representing the public key material of the certificate.
//
// Methods:
// - NewX509Certificate: Creates a new X.509 certificate with the specified subject, key size, and validity period.
// - ExportPFX: Exports the certificate and private key to a PFX file with the specified password.
//
// Note:
// The X509Certificate struct is used to manage X.509 certificates, including the generation of new certificates and the export of certificates and private keys to PFX files.
// The struct includes fields for the RSA private key, X.509 certificate, and public key material. The NewX509Certificate method is used to create a new certificate, and the ExportPFX method is used to export the certificate and private key to a PFX file.
type X509Certificate struct {
	key         *rsa.PrivateKey
	certificate *x509.Certificate
	publicKey   RSAKeyMaterial
}

// NewX509Certificate creates a new X.509 certificate with the specified subject, key size, and validity period.
//
// Parameters:
// - subject: A string representing the common name (CN) of the certificate subject.
// - keySize: An integer specifying the size of the RSA key to be generated (e.g., 2048, 4096).
// - notBefore: A time.Time object representing the start of the certificate's validity period.
// - notAfter: A time.Time object representing the end of the certificate's validity period.
//
// Returns:
// - A pointer to an X509Certificate object containing the generated certificate and associated RSA private key.
// - An error if the certificate generation fails.
//
// Note:
// The function performs the following steps:
// 1. Generates a new RSA private key with the specified key size.
// 2. Creates a serial number for the certificate.
// 3. Constructs a certificate template with the specified subject, validity period, key usage, and extended key usage.
// 4. Creates a self-signed X.509 certificate using the generated RSA private key and certificate template.
// 5. Parses the generated certificate and returns an X509Certificate object containing the certificate and private key.
//
// Example usage:
// cert, err := NewX509Certificate("example.com", 2048, time.Now(), time.Now().AddDate(1, 0, 0))
//
//	if err != nil {
//	    fmt.Printf("Error creating X509Certificate: %s\n", err)
//	}
func NewX509Certificate(subject string, keySize int, notBefore, notAfter time.Time) (*X509Certificate, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: subject,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &rsaKey.PublicKey, rsaKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, err
	}

	return &X509Certificate{
		key:         rsaKey,
		certificate: cert,
	}, nil
}

// Export ====================================================================================

// ExportPFX exports the certificate and private key to a PFX file with the specified password.
//
// Parameters:
// - pathToFile: A string representing the path to the file where the PFX will be exported.
// - password: A string representing the password for the PFX file.
//
// Returns:
// - An error if the export fails, otherwise nil.
func (x *X509Certificate) ExportPFX(pathToFile, password string) error {
	return fmt.Errorf("ExportPFX not implemented")
}

// ExportRSAPublicKeyPEM exports the public key to a PEM file.
//
// Parameters:
// - pathToFile: A string representing the path to the file where the public key will be exported.
//
// Returns:
// - An error if the export fails, otherwise nil.
func (x *X509Certificate) ExportRSAPublicKeyPEM(pathToFile string) error {
	if len(pathToFile) != 0 {
		dir := filepath.Dir(pathToFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}
	}

	pubKeyOut, err := os.Create(pathToFile)
	if err != nil {
		return err
	}
	defer pubKeyOut.Close()

	pubBytes, err := x509.MarshalPKIXPublicKey(&x.key.PublicKey)
	if err != nil {
		return err
	}

	if err := pem.Encode(pubKeyOut, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}); err != nil {
		return err
	}

	return nil
}

// ExportRSAPrivateKeyPEM exports the private key to a PEM file.
//
// Parameters:
// - pathToFile: A string representing the path to the file where the private key will be exported.
//
// Returns:
// - An error if the export fails, otherwise nil.
func (x *X509Certificate) ExportRSAPrivateKeyPEM(pathToFile string) error {
	if len(pathToFile) != 0 {
		dir := filepath.Dir(pathToFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}
	}

	keyOut, err := os.Create(pathToFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privBytes := x509.MarshalPKCS1PrivateKey(x.key)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	return nil
}

// ExportRSAPublicKeyDER exports the public key to a DER file.
//
// Parameters:
// - pathToFile: A string representing the path to the file where the public key will be exported.
//
// Returns:
// - An error if the export fails, otherwise nil.
func (x *X509Certificate) ExportRSAPublicKeyDER() ([]byte, error) {
	return nil, fmt.Errorf("ExportRSAPublicKeyDER not implemented")
}

// ExportRSAPublicKeyBCrypt exports the public key to a BCrypt file.
//
// Parameters:
// - pathToFile: A string representing the path to the file where the public key will be exported.
//
// Returns:
// - An error if the export fails, otherwise nil.
func (x *X509Certificate) ExportRSAPublicKeyBCrypt() ([]byte, error) {
	return nil, fmt.Errorf("ExportRSAPublicKeyBCrypt not implemented")
}

// ExportRSAPublicKey returns the public key material of the certificate.
//
// Returns:
// - An RSAKeyMaterial object representing the public key material of the certificate.
func (x *X509Certificate) ExportRSAPublicKey() RSAKeyMaterial {
	return x.publicKey
}

// GetRSAPublicKey returns the public key of the certificate.
//
// Returns:
// - A pointer to an rsa.PublicKey object representing the public key of the certificate.
func (x *X509Certificate) GetRSAPublicKey() *rsa.PublicKey {
	return &x.key.PublicKey
}

// GetRSAPrivateKey returns the private key of the certificate.
//
// Returns:
// - A pointer to an rsa.PrivateKey object representing the private key of the certificate.
func (x *X509Certificate) GetRSAPrivateKey() *rsa.PrivateKey {
	return x.key
}

// GetCertificate returns the certificate of the certificate.
//
// Returns:
// - A pointer to an x509.Certificate object representing the certificate of the certificate.
func (x *X509Certificate) GetCertificate() *x509.Certificate {
	return x.certificate
}

// GetRSAKeyMaterial returns the RSA key material of the certificate.
//
// Returns:
// - An RSAKeyMaterial object representing the RSA key material of the certificate.
func (x *X509Certificate) GetRSAKeyMaterial() RSAKeyMaterial {
	return RSAKeyMaterial{
		Modulus:  x.key.PublicKey.N.Bytes(),
		Exponent: uint32(x.key.PublicKey.E),
		KeySize:  uint32(x.key.PublicKey.Size() * 8),
		Prime1:   []byte{},
		Prime2:   []byte{},
	}
}
