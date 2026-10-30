package blob

import (
	"bytes"
	"errors"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
)

// BCRYPT_RSA_FULLPRIVATE_BLOB represents the structure used as a header for an RSA private key BLOB in memory.
//
// The layout of a BCRYPT_RSA_FULLPRIVATE_BLOB in memory is as follows:
//
//	BCRYPT_RSA_FULLPRIVATE_BLOB
//	PublicExponent[cbPublicExp] // Big-endian
//	Modulus[cbModulus]         // Big-endian
//	Prime1[cbPrime1] // Big-endian
//	Prime2[cbPrime2] // Big-endian
//	Exponent1[cbPrime1] // Big-endian
//	Exponent2[cbPrime2] // Big-endian
//	Coefficient[cbPrime1] // Big-endian
//	PrivateExponent[cbModulus] // Big-endian
//
// The fields following this structure (PublicExponent, Modulus, Prime1, Prime2, Exponent1, Exponent2, Coefficient, and PrivateExponent) are stored in big-endian format.
//
// Note that in different versions of Windows, the value that PrivateExponent takes from a call of
// BCryptExportKey may be different as there are several mathematically equivalent representations
// of the PrivateExponent in cbModulus bytes. Notably, in some versions the PrivateExponent will be
// exported modulo (Prime1 - 1) * (Prime2 - 1), and in others it will be exported modulo
// LeastCommonMultiple(Prime1 - 1, Prime2 - 1).
//
// See:
// https://learn.microsoft.com/en-us/windows/win32/api/bcrypt/ns-bcrypt-bcrypt_rsaprivate_blob
type BCRYPT_RSA_FULLPRIVATE_BLOB struct {
	// PublicExponent is the public exponent of the RSA key.
	PublicExponent []byte // Big-endian.

	// Modulus is the modulus of the RSA key.
	Modulus []byte // Big-endian.

	// Prime1 is the first prime of the RSA key.
	Prime1 []byte // Big-endian.

	// Prime2 is the second prime of the RSA key.
	Prime2 []byte // Big-endian.

	// Exponent1 is the first exponent of the RSA key.
	Exponent1 []byte // Big-endian.

	// Exponent2 is the second exponent of the RSA key.
	Exponent2 []byte // Big-endian.

	// Coefficient is the coefficient of the RSA key.
	Coefficient []byte // Big-endian.

	// PrivateExponent is the private exponent of the RSA key.
	PrivateExponent []byte // Big-endian.
}

// Unmarshal parses the provided byte slice into the BCRYPT_RSA_FULLPRIVATE_BLOB structure.
//
// Parameters:
// - value: A byte slice containing the raw RSA full private key BLOB to be parsed.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
//
// Note:
// The function expects the byte slice to follow the RSA full private key BLOB format, starting with the BCRYPT_RSA_KEY_BLOB header.
// It extracts the public exponent, modulus, prime1, prime2, exponent1, exponent2, coefficient, and private exponent from the byte slice and stores them in the BCRYPT_RSA_FULLPRIVATE_BLOB structure.
func (b *BCRYPT_RSA_FULLPRIVATE_BLOB) Unmarshal(keyHeader headers.BCRYPT_RSA_KEY_BLOB, value []byte) (int, error) {
	if len(value) < 24 {
		return 0, errors.New("buffer too small for BCRYPT_RSA_FULLPRIVATE_BLOB, header too short (at least 24 bytes are required)")
	}

	// Sizes are compared as uint64: on 32-bit platforms int() of a size
	// >= 2^31 is negative and would slip past a signed bounds check.
	bytesRead := 0

	// Unmarshalling public exponent
	if uint64(keyHeader.CbPublicExp) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_FULLPRIVATE_BLOB, not enough bytes for unmarshalling public exponent")
	}
	b.PublicExponent = value[bytesRead : bytesRead+int(keyHeader.CbPublicExp)]
	bytesRead += int(keyHeader.CbPublicExp)

	// Unmarshalling modulus
	if uint64(keyHeader.CbModulus) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_FULLPRIVATE_BLOB, not enough bytes for unmarshalling modulus")
	}
	b.Modulus = value[bytesRead : bytesRead+int(keyHeader.CbModulus)]
	bytesRead += int(keyHeader.CbModulus)

	// Unmarshalling prime1
	if uint64(keyHeader.CbPrime1) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_FULLPRIVATE_BLOB, not enough bytes for unmarshalling prime1")
	}
	b.Prime1 = value[bytesRead : bytesRead+int(keyHeader.CbPrime1)]
	bytesRead += int(keyHeader.CbPrime1)

	// Unmarshalling prime2
	if uint64(keyHeader.CbPrime2) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_FULLPRIVATE_BLOB, not enough bytes for unmarshalling prime2")
	}
	b.Prime2 = value[bytesRead : bytesRead+int(keyHeader.CbPrime2)]
	bytesRead += int(keyHeader.CbPrime2)

	// Unmarshalling exponent1 (cbPrime1 bytes)
	if uint64(keyHeader.CbPrime1) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_FULLPRIVATE_BLOB, not enough bytes for unmarshalling exponent1")
	}
	b.Exponent1 = value[bytesRead : bytesRead+int(keyHeader.CbPrime1)]
	bytesRead += int(keyHeader.CbPrime1)

	// Unmarshalling exponent2 (cbPrime2 bytes)
	if uint64(keyHeader.CbPrime2) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_FULLPRIVATE_BLOB, not enough bytes for unmarshalling exponent2")
	}
	b.Exponent2 = value[bytesRead : bytesRead+int(keyHeader.CbPrime2)]
	bytesRead += int(keyHeader.CbPrime2)

	// Unmarshalling coefficient (cbPrime1 bytes)
	if uint64(keyHeader.CbPrime1) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_FULLPRIVATE_BLOB, not enough bytes for unmarshalling coefficient")
	}
	b.Coefficient = value[bytesRead : bytesRead+int(keyHeader.CbPrime1)]
	bytesRead += int(keyHeader.CbPrime1)

	// Unmarshalling private exponent (cbModulus bytes)
	if uint64(keyHeader.CbModulus) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_FULLPRIVATE_BLOB, not enough bytes for unmarshalling private exponent")
	}
	b.PrivateExponent = value[bytesRead : bytesRead+int(keyHeader.CbModulus)]
	bytesRead += int(keyHeader.CbModulus)

	return bytesRead, nil
}

// Marshal returns the raw bytes of the BCRYPT_RSA_FULLPRIVATE_BLOB structure.
//
// Returns:
// - A byte slice representing the raw bytes of the BCRYPT_RSA_FULLPRIVATE_BLOB structure.
func (b *BCRYPT_RSA_FULLPRIVATE_BLOB) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	marshalledData = append(marshalledData, b.PublicExponent...)

	marshalledData = append(marshalledData, b.Modulus...)

	marshalledData = append(marshalledData, b.Prime1...)

	marshalledData = append(marshalledData, b.Prime2...)

	marshalledData = append(marshalledData, b.Exponent1...)

	marshalledData = append(marshalledData, b.Exponent2...)

	marshalledData = append(marshalledData, b.Coefficient...)

	marshalledData = append(marshalledData, b.PrivateExponent...)

	return marshalledData, nil
}

// Equal checks if two BCRYPT_RSA_FULLPRIVATE_BLOB structures are equal.
//
// Parameters:
// - other: The BCRYPT_RSA_FULLPRIVATE_BLOB structure to compare to.
//
// Returns:
// - True if the two BCRYPT_RSA_FULLPRIVATE_BLOB structures are equal, false otherwise.
func (b *BCRYPT_RSA_FULLPRIVATE_BLOB) Equal(other *BCRYPT_RSA_FULLPRIVATE_BLOB) bool {
	return bytes.Equal(b.PublicExponent, other.PublicExponent) && bytes.Equal(b.Modulus, other.Modulus) && bytes.Equal(b.Prime1, other.Prime1) && bytes.Equal(b.Prime2, other.Prime2) && bytes.Equal(b.Exponent1, other.Exponent1) && bytes.Equal(b.Exponent2, other.Exponent2) && bytes.Equal(b.Coefficient, other.Coefficient) && bytes.Equal(b.PrivateExponent, other.PrivateExponent)
}
