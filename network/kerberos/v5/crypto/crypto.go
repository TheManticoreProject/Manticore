// Package kerbcrypto provides Kerberos cryptographic operations including
// string-to-key derivation, encryption, and decryption for RC4-HMAC and
// AES-CTS-HMAC-SHA1-96 encryption types.
//
// Import path: github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto
package kerbcrypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// Key usage constants per RFC 4120 Section 7.5.1, re-exported from the iana
// leaf package so callers can keep using kerbcrypto.KeyUsage* while iana
// remains the single source of truth.
const (
	KeyUsageASReqPAEncTimestamp    = iana.KeyUsageASReqPAEncTimestamp
	KeyUsageKDCRepTicket           = iana.KeyUsageKDCRepTicket
	KeyUsageASRepEncPart           = iana.KeyUsageASRepEncPart
	KeyUsageTGSReqPAAPReqAuthen    = iana.KeyUsageTGSReqPAAPReqAuthen
	KeyUsageTGSRepEncSessionKey    = iana.KeyUsageTGSRepEncSessionKey
	KeyUsageTGSRepEncSubSessionKey = iana.KeyUsageTGSRepEncSubSessionKey
	KeyUsageAPReqAuthen            = iana.KeyUsageAPReqAuthen
	KeyUsageAPRepEncPart           = iana.KeyUsageAPRepEncPart
	KeyUsageKRBCredEncPart         = iana.KeyUsageKRBCredEncPart
	KeyUsageKerbNonKerbSalt        = iana.KeyUsageKerbNonKerbSalt
)

// Sentinel errors for cryptographic operations.
var (
	// ErrCiphertextTooShort is returned when the ciphertext is too short to be valid.
	ErrCiphertextTooShort = errors.New("kerbcrypto: ciphertext too short")
	// ErrIntegrityCheckFailed is returned when the MAC verification fails.
	ErrIntegrityCheckFailed = errors.New("kerbcrypto: integrity check failed")
	// ErrUnsupportedEType is returned when an encryption type is not supported.
	ErrUnsupportedEType = errors.New("kerbcrypto: unsupported encryption type")
)

// randRead fills buf with cryptographically random bytes.
// It wraps crypto/rand.Read as a package-level variable for testability.
var randRead = func(buf []byte) (int, error) {
	return io.ReadFull(rand.Reader, buf)
}

// StringToKey derives an encryption key from a password and salt for the given etype.
// For RC4-HMAC (etype 23), the salt is ignored.
// For AES (etype 17/18), the salt is used with PBKDF2-HMAC-SHA1.
// The params argument carries S2KParams from PA-ETYPE-INFO2 (currently only iteration count
// for AES is supported; pass nil for defaults).
func StringToKey(etype int, password, salt string, params []byte) ([]byte, error) {
	switch etype {
	case iana.ETypeRC4HMAC:
		// RC4-HMAC: key = NT hash of password; salt is not used
		return rc4HMACStringToKey(password), nil

	case iana.ETypeAES128CTSHMACSHA196:
		iter_count := aesDefaultIterCount
		if len(params) >= 4 {
			// S2KParams contains a 4-byte big-endian iteration count
			iter_count = int(params[0])<<24 | int(params[1])<<16 | int(params[2])<<8 | int(params[3])
			if iter_count <= 0 {
				iter_count = aesDefaultIterCount
			}
		}
		return aesStringToKey(password, salt, iter_count, 16)

	case iana.ETypeAES256CTSHMACSHA196:
		iter_count := aesDefaultIterCount
		if len(params) >= 4 {
			iter_count = int(params[0])<<24 | int(params[1])<<16 | int(params[2])<<8 | int(params[3])
			if iter_count <= 0 {
				iter_count = aesDefaultIterCount
			}
		}
		return aesStringToKey(password, salt, iter_count, 32)

	case iana.ETypeAES128CTSHMACSHA256, iana.ETypeAES256CTSHMACSHA384:
		p, _ := aes2ParamsFor(etype)
		iter_count := aes8009DefaultIterCount
		if len(params) >= 4 {
			iter_count = int(params[0])<<24 | int(params[1])<<16 | int(params[2])<<8 | int(params[3])
			if iter_count <= 0 {
				iter_count = aes8009DefaultIterCount
			}
		}
		return aes2StringToKey(password, salt, iter_count, p)

	default:
		return nil, fmt.Errorf("%w: %d", ErrUnsupportedEType, etype)
	}
}

// Encrypt encrypts plaintext with the given key, etype, and key usage number.
// Returns the ciphertext including confounder and MAC.
func Encrypt(etype int, key []byte, usage int, plaintext []byte) ([]byte, error) {
	switch etype {
	case iana.ETypeRC4HMAC:
		return rc4HMACEncrypt(key, usage, plaintext)
	case iana.ETypeAES128CTSHMACSHA196, iana.ETypeAES256CTSHMACSHA196:
		return aesEncrypt(key, etype, usage, plaintext)
	case iana.ETypeAES128CTSHMACSHA256, iana.ETypeAES256CTSHMACSHA384:
		return aes2Encrypt(key, etype, usage, plaintext)
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnsupportedEType, etype)
	}
}

// Decrypt decrypts ciphertext with the given key, etype, and key usage number.
// Returns the plaintext (confounder is stripped).
func Decrypt(etype int, key []byte, usage int, ciphertext []byte) ([]byte, error) {
	switch etype {
	case iana.ETypeRC4HMAC:
		return rc4HMACDecrypt(key, usage, ciphertext)
	case iana.ETypeAES128CTSHMACSHA196, iana.ETypeAES256CTSHMACSHA196:
		return aesDecrypt(key, etype, usage, ciphertext)
	case iana.ETypeAES128CTSHMACSHA256, iana.ETypeAES256CTSHMACSHA384:
		return aes2Decrypt(key, etype, usage, ciphertext)
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnsupportedEType, etype)
	}
}

// KeyLen returns the key length in bytes for the given etype.
func KeyLen(etype int) int {
	switch etype {
	case iana.ETypeRC4HMAC:
		return 16
	case iana.ETypeAES128CTSHMACSHA196:
		return 16
	case iana.ETypeAES256CTSHMACSHA196:
		return 32
	case iana.ETypeAES128CTSHMACSHA256:
		return 16
	case iana.ETypeAES256CTSHMACSHA384:
		return 32
	default:
		return 0
	}
}
