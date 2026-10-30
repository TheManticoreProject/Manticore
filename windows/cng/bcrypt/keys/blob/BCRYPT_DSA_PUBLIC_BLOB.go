package blob

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
)

// BCRYPT_DSA_PUBLIC_BLOB represents the structure used for a DSA public key BLOB in memory.
//
// The layout of a BCRYPT_DSA_PUBLIC_BLOB in memory is as follows:
//
//	BCRYPT_DSA_PUBLIC_BLOB
//	Modulus[cbKey]    // Big-endian
//	Generator[cbKey]  // Big-endian
//	Public[cbKey]     // Big-endian
//
// The fields following this structure are stored in big-endian format.
//
// See:
// https://learn.microsoft.com/en-us/windows/win32/api/bcrypt/ns-bcrypt-bcrypt_dsa_key_blob
type BCRYPT_DSA_PUBLIC_BLOB struct {
	// Modulus is the Modulus (P) parameter of the DSA key.
	Modulus []byte // Big-endian.

	// Generator is the Generator (G) parameter of the DSA key.
	Generator []byte // Big-endian.

	// Public is the public value (Y) of the DSA key.
	Public []byte // Big-endian.
}

// Unmarshal parses the provided byte slice into the BCRYPT_DSA_PUBLIC_BLOB structure.
//
// Parameters:
// - value: A byte slice containing the raw DSA key BLOB to be parsed.
// - keyHeader: A struct containing the cbKey field for size.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
func (b *BCRYPT_DSA_PUBLIC_BLOB) Unmarshal(keyHeader headers.BCRYPT_DSA_KEY_BLOB, value []byte) (int, error) {
	bytesRead := 0
	// On 32-bit platforms int() of a size >= 2^31 is negative, so each
	// bounds check below also rejects negative values.
	keyLen := int(keyHeader.CbKey)

	if keyLen < 0 || keyLen > len(value)-bytesRead {
		return 0, errors.New("buffer too small for BCRYPT_DSA_PUBLIC_BLOB, not enough bytes for unmarshalling modulus")
	}
	b.Modulus = value[bytesRead : bytesRead+keyLen]
	bytesRead += keyLen

	if keyLen < 0 || keyLen > len(value)-bytesRead {
		return 0, errors.New("buffer too small for BCRYPT_DSA_PUBLIC_BLOB, not enough bytes for unmarshalling generator")
	}
	b.Generator = value[bytesRead : bytesRead+keyLen]
	bytesRead += keyLen

	if keyLen < 0 || keyLen > len(value)-bytesRead {
		return 0, errors.New("buffer too small for BCRYPT_DSA_PUBLIC_BLOB, not enough bytes for unmarshalling public value")
	}
	b.Public = value[bytesRead : bytesRead+keyLen]
	bytesRead += keyLen

	return bytesRead, nil
}

// Marshal returns the raw bytes of the BCRYPT_DSA_PUBLIC_BLOB structure.
//
// Returns:
// - A byte slice representing the raw bytes of the BCRYPT_DSA_PUBLIC_BLOB structure.
func (b *BCRYPT_DSA_PUBLIC_BLOB) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	marshalledData = append(marshalledData, b.Modulus...)
	marshalledData = append(marshalledData, b.Generator...)
	marshalledData = append(marshalledData, b.Public...)

	return marshalledData, nil
}

// Describe prints the BCRYPT_DSA_PUBLIC_BLOB structure to the console.
//
// Parameters:
// - indent: The number of spaces to indent the output.
func (b *BCRYPT_DSA_PUBLIC_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mBCRYPT_DSA_PRIVATE_BLOB\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mModulus\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(b.Modulus))
	fmt.Printf("%s │ \x1b[93mGenerator\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(b.Generator))
	fmt.Printf("%s │ \x1b[93mPublic\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(b.Public))
	fmt.Printf("%s └───\n", indentPrompt)
}

// Equal checks if two BCRYPT_DSA_PUBLIC_BLOB structures are equal.
//
// Parameters:
// - other: The BCRYPT_DSA_PUBLIC_BLOB structure to compare to.
//
// Returns:
// - True if the two BCRYPT_DSA_PUBLIC_BLOB structures are equal, false otherwise.
func (b *BCRYPT_DSA_PUBLIC_BLOB) Equal(other *BCRYPT_DSA_PUBLIC_BLOB) bool {
	return bytes.Equal(b.Modulus, other.Modulus) && bytes.Equal(b.Generator, other.Generator) && bytes.Equal(b.Public, other.Public)
}
