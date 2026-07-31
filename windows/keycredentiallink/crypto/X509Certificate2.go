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

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/blob"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/magic"

	"software.sslmate.com/src/go-pkcs12"
)

const (
	// privateKeyFileMode is the mode of a file holding private key material. Only
	// the owner may read or write it.
	privateKeyFileMode = 0o600

	// privateKeyDirMode is the mode of a directory created to hold private key
	// material. Only the owner may traverse or list it.
	privateKeyDirMode = 0o700
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
	privateKey  *rsa.PrivateKey
	certificate *x509.Certificate
	publicKey   *rsa.PublicKey
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
		privateKey:  rsaKey,
		publicKey:   &rsaKey.PublicKey,
		certificate: cert,
	}, nil
}

// Export ====================================================================================

// ExportCertificatePEM exports the certificate to a PEM file.
//
// This is the counterpart of ExportRSAPrivateKeyPEM: together the two files are
// what PKINIT implementations expect when they take a certificate and its private
// key as separate PEM inputs. The directory is created with privateKeyDirMode
// because it is the same directory the private key is written to.
//
// Parameters:
// - pathToFile: A string representing the path to the file where the certificate will be exported.
//
// Returns:
// - An error if the export fails, otherwise nil.
func (x *X509Certificate) ExportCertificatePEM(pathToFile string) error {
	certificate, err := x.GetCertificate()
	if err != nil {
		return fmt.Errorf("error getting certificate from crypto.X509Certificate: %s", err)
	}

	if len(pathToFile) != 0 {
		dir := filepath.Dir(pathToFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, privateKeyDirMode); err != nil {
				return err
			}
		}
	}

	certOut, err := os.Create(pathToFile)
	if err != nil {
		return err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw}); err != nil {
		return err
	}

	return nil
}

// ExportPFX exports the certificate and private key to a PFX file with the specified password.
//
// The PFX is encoded with the modern algorithm set (AES-256-CBC and SHA-256 for
// both encryption and the integrity MAC), which is what current OpenSSL, .NET and
// Python implementations read. The legacy RC2 and DES encodings are deliberately
// not used, as modern OpenSSL rejects them by default.
//
// The PFX embeds the private key, so the file is written with privateKeyFileMode
// and a directory created to hold it with privateKeyDirMode.
//
// Parameters:
// - pathToFile: A string representing the path to the file where the PFX will be exported.
// - password: A string representing the password for the PFX file.
//
// Returns:
// - An error if the export fails, otherwise nil.
func (x *X509Certificate) ExportPFX(pathToFile, password string) error {
	privateKey, err := x.GetRSAPrivateKey()
	if err != nil {
		return fmt.Errorf("error getting RSA private key from crypto.X509Certificate: %s", err)
	}

	certificate, err := x.GetCertificate()
	if err != nil {
		return fmt.Errorf("error getting certificate from crypto.X509Certificate: %s", err)
	}

	pfxData, err := pkcs12.Modern.Encode(privateKey, certificate, nil, password)
	if err != nil {
		return fmt.Errorf("error encoding PFX: %s", err)
	}

	if len(pathToFile) != 0 {
		dir := filepath.Dir(pathToFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, privateKeyDirMode); err != nil {
				return err
			}
		}
	}

	pfxOut, err := os.OpenFile(pathToFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, privateKeyFileMode)
	if err != nil {
		return err
	}
	defer pfxOut.Close()

	// O_CREATE leaves the mode of an already existing file untouched, so narrow it
	// explicitly before the key material is written.
	if err := pfxOut.Chmod(privateKeyFileMode); err != nil {
		return err
	}

	if _, err := pfxOut.Write(pfxData); err != nil {
		return err
	}

	return nil
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

	publicKey, err := x.GetRSAPublicKey()
	if err != nil {
		return fmt.Errorf("error getting RSA public key from crypto.X509Certificate: %s", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}

	if err := pem.Encode(pubKeyOut, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}); err != nil {
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

// ExportRSAPublicKeyBCrypt exports the public key to a BCrypt structure.
//
// Parameters:
// - None
//
// Returns:
// - A pointer to a BCRYPT_RSA_PUBLIC_KEY object representing the public key.
// - An error if the export fails, otherwise nil.
func (x *X509Certificate) ExportRSAPublicKeyBCrypt() (*keys.BCRYPT_RSA_PUBLIC_KEY, error) {
	publicKey, err := x.GetRSAPublicKey()
	if err != nil {
		return nil, fmt.Errorf("error getting RSA public key from crypto.X509Certificate: %s", err)
	}

	exponentBigInt := big.NewInt(int64(publicKey.E))
	exponentBytes := exponentBigInt.Bytes()

	return &keys.BCRYPT_RSA_PUBLIC_KEY{
		Magic: magic.BCRYPT_KEY_BLOB{Magic: magic.BCRYPT_RSAPUBLIC_MAGIC},
		Header: headers.BCRYPT_RSA_KEY_BLOB{
			BitLength:   uint32(publicKey.Size() * 8),
			CbPublicExp: uint32(len(exponentBytes)),
			CbModulus:   uint32(len(publicKey.N.Bytes())),
		},
		Content: blob.BCRYPT_RSA_PUBLIC_BLOB{
			PublicExponent: exponentBytes,
			Modulus:        publicKey.N.Bytes(),
		},
	}, nil
}

// ExportRSAPrivateKeyPEM exports the private key to a PEM file.
//
// The file is created with mode 0600 and a directory created to hold it with mode
// 0700, so that the private key is only readable by its owner. An existing file at
// pathToFile is narrowed to 0600 as well, since os.OpenFile leaves the mode of an
// already existing file untouched.
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
			if err := os.MkdirAll(dir, privateKeyDirMode); err != nil {
				return err
			}
		}
	}

	keyOut, err := os.OpenFile(pathToFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, privateKeyFileMode)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	if err := keyOut.Chmod(privateKeyFileMode); err != nil {
		return err
	}

	privateKey, err := x.GetRSAPrivateKey()
	if err != nil {
		return fmt.Errorf("error getting RSA private key from crypto.X509Certificate: %s", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	return nil
}

// ExportRSAPrivateKeyBCrypt exports the private key to a BCrypt file.
//
// Parameters:
// - pathToFile: A string representing the path to the file where the private key will be exported.
//
// Returns:
// - An error if the export fails, otherwise nil.
// - A pointer to a BCRYPT_RSA_PRIVATE_KEY object representing the private key.
func (x *X509Certificate) ExportRSAPrivateKeyBCrypt() (*keys.BCRYPT_RSA_PRIVATE_KEY, error) {
	privateKey, err := x.GetRSAPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("error getting RSA private key from crypto.X509Certificate: %s", err)
	}

	exponentBigInt := big.NewInt(int64(privateKey.E))
	exponentBytes := exponentBigInt.Bytes()

	return &keys.BCRYPT_RSA_PRIVATE_KEY{
		Magic: magic.BCRYPT_KEY_BLOB{Magic: magic.BCRYPT_RSAPRIVATE_MAGIC},
		Header: headers.BCRYPT_RSA_KEY_BLOB{
			BitLength:   uint32(privateKey.PublicKey.Size() * 8),
			CbPublicExp: uint32(len(exponentBytes)),
			CbModulus:   uint32(len(privateKey.PublicKey.N.Bytes())),
			CbPrime1:    uint32(len(privateKey.Primes[0].Bytes())),
			CbPrime2:    uint32(len(privateKey.Primes[1].Bytes())),
		},
		Content: blob.BCRYPT_RSA_PRIVATE_BLOB{
			PublicExponent: exponentBytes,
			Modulus:        privateKey.PublicKey.N.Bytes(),
			Prime1:         privateKey.Primes[0].Bytes(),
			Prime2:         privateKey.Primes[1].Bytes(),
		},
	}, nil
}

// GetRSAPublicKey returns the public key of the certificate.
//
// Returns:
// - A pointer to an rsa.PublicKey object representing the public key of the certificate.
// - An error if the public key retrieval fails, otherwise nil.
func (x *X509Certificate) GetRSAPublicKey() (*rsa.PublicKey, error) {
	if x.publicKey == nil {
		return nil, fmt.Errorf("error getting RSA public key from crypto.X509Certificate, publicKey is nil")
	}
	return x.publicKey, nil
}

// GetRSAPrivateKey returns the private key of the certificate.
//
// Returns:
// - A pointer to an rsa.PrivateKey object representing the private key of the certificate.
// - An error if the private key retrieval fails, otherwise nil.
func (x *X509Certificate) GetRSAPrivateKey() (*rsa.PrivateKey, error) {
	if x.privateKey == nil {
		return nil, fmt.Errorf("error getting RSA private key from crypto.X509Certificate, privateKey is nil")
	}
	return x.privateKey, nil
}

// GetCertificate returns the certificate of the certificate.
//
// Returns:
// - A pointer to an x509.Certificate object representing the certificate of the certificate.
// - An error if the certificate retrieval fails, otherwise nil.
func (x *X509Certificate) GetCertificate() (*x509.Certificate, error) {
	if x.certificate == nil {
		return nil, fmt.Errorf("error getting certificate from crypto.X509Certificate, certificate is nil")
	}
	return x.certificate, nil
}
