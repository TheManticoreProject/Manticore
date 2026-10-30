package blob

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
)

// BCRYPT_RSA_PUBLIC_BLOB represents the structure used as a header for an RSA public key BLOB in memory.
//
// The layout of a BCRYPT_RSA_PUBLIC_BLOB in memory is as follows:
//
//	BCRYPT_RSA_PUBLIC_BLOB
//	PublicExponent[cbPublicExp] // Big-endian
//	Modulus[cbModulus]         // Big-endian
//
// The fields following this structure (PublicExponent and Modulus) are stored in big-endian format.
//
// See:
// https://learn.microsoft.com/en-us/windows/win32/api/bcrypt/ns-bcrypt-bcrypt_rsakey_blob
type BCRYPT_RSA_PUBLIC_BLOB struct {
	// PublicExponent is the public exponent of the RSA key.
	// Stored in big-endian format.
	PublicExponent []byte

	// Modulus is the modulus of the RSA key.
	// Stored in big-endian format.
	Modulus []byte
}

// Unmarshal parses the provided byte slice into the BCRYPT_RSA_PUBLIC_BLOB structure.
//
// Parameters:
// - value: A byte slice containing the raw RSA public key BLOB to be parsed.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
//
// Note:
// The function expects the byte slice to follow the RSA public key BLOB format, starting with the BCRYPT_RSA_KEY_BLOB header.
// It extracts the public exponent and modulus from the byte slice and stores them in the BCRYPT_RSA_PUBLIC_BLOB structure.
func (b *BCRYPT_RSA_PUBLIC_BLOB) Unmarshal(keyHeader headers.BCRYPT_RSA_KEY_BLOB, value []byte) (int, error) {
	// Sizes are compared as uint64: on 32-bit platforms int() of a size
	// >= 2^31 is negative and would slip past a signed bounds check.
	bytesRead := 0

	if uint64(keyHeader.CbPublicExp) > uint64(len(value)-bytesRead) {
		return 0, fmt.Errorf("buffer too small for BCRYPT_RSA_PUBLIC_BLOB, not enough bytes for unmarshalling public exponent")
	}
	b.PublicExponent = value[bytesRead : bytesRead+int(keyHeader.CbPublicExp)]
	bytesRead += int(keyHeader.CbPublicExp)

	if uint64(keyHeader.CbModulus) > uint64(len(value)-bytesRead) {
		return 0, fmt.Errorf("buffer too small for BCRYPT_RSA_PUBLIC_BLOB, not enough bytes for unmarshalling modulus")
	}
	b.Modulus = value[bytesRead : bytesRead+int(keyHeader.CbModulus)]
	bytesRead += int(keyHeader.CbModulus)

	return bytesRead, nil
}

// Marshal returns the raw bytes of the BCRYPT_RSA_PUBLIC_BLOB structure.
//
// Returns:
// - A byte slice representing the raw bytes of the BCRYPT_RSA_PUBLIC_BLOB structure.
func (b *BCRYPT_RSA_PUBLIC_BLOB) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	marshalledData = append(marshalledData, b.PublicExponent...)

	marshalledData = append(marshalledData, b.Modulus...)

	return marshalledData, nil
}

// Describe prints the BCRYPT_RSA_PUBLIC_BLOB structure to the console.
//
// Parameters:
// - indent: The number of spaces to indent the output.
func (b *BCRYPT_RSA_PUBLIC_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mBCRYPT_RSA_PUBLIC_BLOB (content)\x1b[0m>\n", indentPrompt)
	bigIntExponent := big.NewInt(0).SetBytes(b.PublicExponent)
	fmt.Printf("%s │ \x1b[93mPublicExponent\x1b[0m: 0x%x (%d)\n", indentPrompt, b.PublicExponent, bigIntExponent.Int64())
	fmt.Printf("%s │ \x1b[93mModulus\x1b[0m: 0x%s\n", indentPrompt, hex.EncodeToString(b.Modulus))
	fmt.Printf("%s └───\n", indentPrompt)
}

// Equal checks if two BCRYPT_RSA_PUBLIC_BLOB structures are equal.
//
// Parameters:
// - other: The BCRYPT_RSA_PUBLIC_BLOB structure to compare to.
//
// Returns:
// - True if the two BCRYPT_RSA_PUBLIC_BLOB structures are equal, false otherwise.
func (b *BCRYPT_RSA_PUBLIC_BLOB) Equal(other *BCRYPT_RSA_PUBLIC_BLOB) bool {
	return bytes.Equal(b.PublicExponent, other.PublicExponent) && bytes.Equal(b.Modulus, other.Modulus)
}
