package blob

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
)

// BCRYPT_ECC_PUBLIC_BLOB represents the content of an ECC public key BLOB in memory.
//
// Layout in memory (following the BCRYPT_ECC_KEY_BLOB header):
//
//	BCRYPT_ECC_KEY_BLOB
//	BYTE X[cbKey] // Big-endian.
//	BYTE Y[cbKey] // Big-endian.
//
// See:
// https://learn.microsoft.com/en-us/windows/win32/api/bcrypt/ns-bcrypt-bcrypt_ecckey_blob
type BCRYPT_ECC_PUBLIC_BLOB struct {
	// X is the X coordinate of the public point. Big-endian.
	X []byte

	// Y is the Y coordinate of the public point. Big-endian.
	Y []byte
}

// Unmarshal parses the provided byte slice into the BCRYPT_ECC_PUBLIC_BLOB structure.
//
// Parameters:
//   - keyHeader: The already-parsed BCRYPT_ECC_KEY_BLOB header, providing cbKey.
//   - value: A byte slice containing the raw ECC public key BLOB content to be parsed,
//     starting immediately after the header.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
func (b *BCRYPT_ECC_PUBLIC_BLOB) Unmarshal(keyHeader headers.BCRYPT_ECC_KEY_BLOB, value []byte) (int, error) {
	bytesRead := 0
	// On 32-bit platforms int() of a size >= 2^31 is negative, so each
	// bounds check below also rejects negative values.
	cbKey := int(keyHeader.KeySize)

	if cbKey < 0 || cbKey > len(value)-bytesRead {
		return 0, fmt.Errorf("buffer too small for BCRYPT_ECC_PUBLIC_BLOB, not enough bytes for unmarshalling X")
	}
	b.X = value[bytesRead : bytesRead+cbKey]
	bytesRead += cbKey

	if cbKey < 0 || cbKey > len(value)-bytesRead {
		return 0, fmt.Errorf("buffer too small for BCRYPT_ECC_PUBLIC_BLOB, not enough bytes for unmarshalling Y")
	}
	b.Y = value[bytesRead : bytesRead+cbKey]
	bytesRead += cbKey

	return bytesRead, nil
}

// Marshal returns the raw bytes of the BCRYPT_ECC_PUBLIC_BLOB structure.
//
// Returns:
// - A byte slice representing the raw bytes of the BCRYPT_ECC_PUBLIC_BLOB structure.
func (b *BCRYPT_ECC_PUBLIC_BLOB) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	marshalledData = append(marshalledData, b.X...)
	marshalledData = append(marshalledData, b.Y...)

	return marshalledData, nil
}

// Describe prints the BCRYPT_ECC_PUBLIC_BLOB structure to the console.
//
// Parameters:
// - indent: The number of spaces to indent the output.
func (b *BCRYPT_ECC_PUBLIC_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mBCRYPT_ECC_PUBLIC_BLOB (content)\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mX\x1b[0m: 0x%s\n", indentPrompt, hex.EncodeToString(b.X))
	fmt.Printf("%s │ \x1b[93mY\x1b[0m: 0x%s\n", indentPrompt, hex.EncodeToString(b.Y))
	fmt.Printf("%s └───\n", indentPrompt)
}

// Equal checks if two BCRYPT_ECC_PUBLIC_BLOB structures are equal.
//
// Parameters:
// - other: The BCRYPT_ECC_PUBLIC_BLOB structure to compare to.
//
// Returns:
// - True if the two BCRYPT_ECC_PUBLIC_BLOB structures are equal, false otherwise.
func (b *BCRYPT_ECC_PUBLIC_BLOB) Equal(other *BCRYPT_ECC_PUBLIC_BLOB) bool {
	return bytes.Equal(b.X, other.X) && bytes.Equal(b.Y, other.Y)
}
