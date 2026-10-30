package blob

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
)

// BCRYPT_RSA_PRIVATE_BLOB represents the structure used as a header for an RSA private key BLOB in memory.
//
// The layout of a BCRYPT_RSA_PRIVATE_BLOB in memory is as follows:
//
//	BCRYPT_RSA_PRIVATE_BLOB
//	PublicExponent[cbPublicExp] // Big-endian
//	Modulus[cbModulus]         // Big-endian
//	Prime1[cbPrime1] // Big-endian
//	Prime2[cbPrime2] // Big-endian
//
// The fields following this structure (PublicExponent, Modulus, Prime1, and Prime2) are stored in big-endian format.
//
// See:
// https://learn.microsoft.com/en-us/windows/win32/api/bcrypt/ns-bcrypt-bcrypt_rsakey_blob
type BCRYPT_RSA_PRIVATE_BLOB struct {
	// PublicExponent is the public exponent of the RSA key.
	PublicExponent []byte // Big-endian.

	// Modulus is the modulus of the RSA key.
	Modulus []byte // Big-endian.

	// Prime1 is the first prime of the RSA key.
	Prime1 []byte // Big-endian.

	// Prime2 is the second prime of the RSA key.
	Prime2 []byte // Big-endian.
}

// Unmarshal parses the provided byte slice into the BCRYPT_RSA_PRIVATE_BLOB structure.
//
// Parameters:
// - value: A byte slice containing the raw RSA private key BLOB to be parsed.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
//
// Note:
// The function expects the byte slice to follow the RSA private key BLOB format, starting with the BCRYPT_RSA_KEY_BLOB header.
// It extracts the public exponent, modulus, prime1, and prime2 from the byte slice and stores them in the BCRYPT_RSA_PRIVATE_BLOB structure.
func (b *BCRYPT_RSA_PRIVATE_BLOB) Unmarshal(keyHeader headers.BCRYPT_RSA_KEY_BLOB, value []byte) (int, error) {
	// Sizes are compared as uint64: on 32-bit platforms int() of a size
	// >= 2^31 is negative and would slip past a signed bounds check.
	bytesRead := 0

	if uint64(keyHeader.CbPublicExp) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_PRIVATE_BLOB, not enough bytes for unmarshalling public exponent")
	}
	b.PublicExponent = value[bytesRead : bytesRead+int(keyHeader.CbPublicExp)]
	bytesRead += int(keyHeader.CbPublicExp)

	if uint64(keyHeader.CbModulus) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_PRIVATE_BLOB, not enough bytes for unmarshalling modulus")
	}
	b.Modulus = value[bytesRead : bytesRead+int(keyHeader.CbModulus)]
	bytesRead += int(keyHeader.CbModulus)

	if uint64(keyHeader.CbPrime1) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_PRIVATE_BLOB, not enough bytes for unmarshalling prime1")
	}
	b.Prime1 = value[bytesRead : bytesRead+int(keyHeader.CbPrime1)]
	bytesRead += int(keyHeader.CbPrime1)

	if uint64(keyHeader.CbPrime2) > uint64(len(value)-bytesRead) {
		return 0, errors.New("buffer too small for BCRYPT_RSA_PRIVATE_BLOB, not enough bytes for unmarshalling prime2")
	}
	b.Prime2 = value[bytesRead : bytesRead+int(keyHeader.CbPrime2)]
	bytesRead += int(keyHeader.CbPrime2)

	return bytesRead, nil
}

// Marshal returns the raw bytes of the BCRYPT_RSA_PRIVATE_BLOB structure.
//
// Returns:
// - A byte slice representing the raw bytes of the BCRYPT_RSA_PRIVATE_BLOB structure.
func (b *BCRYPT_RSA_PRIVATE_BLOB) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	marshalledData = append(marshalledData, b.PublicExponent...)

	marshalledData = append(marshalledData, b.Modulus...)

	marshalledData = append(marshalledData, b.Prime1...)

	marshalledData = append(marshalledData, b.Prime2...)

	return marshalledData, nil
}

// Describe prints the BCRYPT_RSA_PRIVATE_BLOB structure to the console.
//
// Parameters:
// - indent: The number of spaces to indent the output.
func (b *BCRYPT_RSA_PRIVATE_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mBCRYPT_RSA_PRIVATE_BLOB\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mPublicExponent\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(b.PublicExponent))
	fmt.Printf("%s │ \x1b[93mModulus\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(b.Modulus))
	fmt.Printf("%s │ \x1b[93mPrime1\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(b.Prime1))
	fmt.Printf("%s │ \x1b[93mPrime2\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(b.Prime2))
	fmt.Printf("%s └───\n", indentPrompt)
}

// Equal checks if two BCRYPT_RSA_PRIVATE_BLOB structures are equal.
//
// Parameters:
// - other: The BCRYPT_RSA_PRIVATE_BLOB structure to compare to.
//
// Returns:
// - True if the two BCRYPT_RSA_PRIVATE_BLOB structures are equal, false otherwise.
func (b *BCRYPT_RSA_PRIVATE_BLOB) Equal(other *BCRYPT_RSA_PRIVATE_BLOB) bool {
	return bytes.Equal(b.PublicExponent, other.PublicExponent) && bytes.Equal(b.Modulus, other.Modulus) && bytes.Equal(b.Prime1, other.Prime1) && bytes.Equal(b.Prime2, other.Prime2)
}
