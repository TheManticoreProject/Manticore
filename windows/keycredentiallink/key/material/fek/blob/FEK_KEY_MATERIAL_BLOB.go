package blob

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/headers"
)

// FEK_KEY_MATERIAL_BLOB carries the variable-length fields of the KEY_USAGE_FEK
// key material: the RSA 2048 public exponent, the RSA 2048 modulus, and the
// AES-256 KDF key. The sizes of each field are supplied by FEK_KEY_MATERIAL_HEADER.
//
// The layout of a FEK_KEY_MATERIAL_BLOB in memory is as follows:
//
//	PublicExponent[CbPublicExp] // Big-endian
//	Modulus[CbModulus]          // Big-endian
//	AESKDFKey[CbAESKDFKey]      // Raw bytes
//
// See:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/735fd27a-3f22-4926-93f9-0298bb67a84b
type FEK_KEY_MATERIAL_BLOB struct {
	// PublicExponent is the public exponent of the RSA 2048 key, in big-endian.
	PublicExponent []byte

	// Modulus is the modulus of the RSA 2048 key, in big-endian.
	Modulus []byte

	// AESKDFKey is the AES-256 KDF key material.
	AESKDFKey []byte
}

// Unmarshal parses the provided byte slice into the FEK_KEY_MATERIAL_BLOB structure.
//
// Parameters:
// - keyHeader: The FEK_KEY_MATERIAL_HEADER describing the field sizes.
// - value: A byte slice containing the raw FEK key material blob to be parsed.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
func (b *FEK_KEY_MATERIAL_BLOB) Unmarshal(keyHeader headers.FEK_KEY_MATERIAL_HEADER, value []byte) (int, error) {
	bytesRead := 0

	if int(keyHeader.CbPublicExp) > len(value)-bytesRead {
		return 0, fmt.Errorf("buffer too small for FEK_KEY_MATERIAL_BLOB, not enough bytes for unmarshalling public exponent")
	}
	b.PublicExponent = value[bytesRead : bytesRead+int(keyHeader.CbPublicExp)]
	bytesRead += int(keyHeader.CbPublicExp)

	if int(keyHeader.CbModulus) > len(value)-bytesRead {
		return 0, fmt.Errorf("buffer too small for FEK_KEY_MATERIAL_BLOB, not enough bytes for unmarshalling modulus")
	}
	b.Modulus = value[bytesRead : bytesRead+int(keyHeader.CbModulus)]
	bytesRead += int(keyHeader.CbModulus)

	if int(keyHeader.CbAESKDFKey) > len(value)-bytesRead {
		return 0, fmt.Errorf("buffer too small for FEK_KEY_MATERIAL_BLOB, not enough bytes for unmarshalling AES-256 KDF key")
	}
	b.AESKDFKey = value[bytesRead : bytesRead+int(keyHeader.CbAESKDFKey)]
	bytesRead += int(keyHeader.CbAESKDFKey)

	return bytesRead, nil
}

// Marshal returns the raw bytes of the FEK_KEY_MATERIAL_BLOB structure.
//
// Returns:
// - A byte slice representing the raw bytes of the FEK_KEY_MATERIAL_BLOB structure.
func (b *FEK_KEY_MATERIAL_BLOB) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	marshalledData = append(marshalledData, b.PublicExponent...)

	marshalledData = append(marshalledData, b.Modulus...)

	marshalledData = append(marshalledData, b.AESKDFKey...)

	return marshalledData, nil
}

// Describe prints the FEK_KEY_MATERIAL_BLOB structure to the console.
//
// Parameters:
// - indent: The number of levels to indent the output.
func (b *FEK_KEY_MATERIAL_BLOB) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mFEK_KEY_MATERIAL_BLOB (content)\x1b[0m>\n", indentPrompt)
	bigIntExponent := big.NewInt(0).SetBytes(b.PublicExponent)
	fmt.Printf("%s │ \x1b[93mPublicExponent\x1b[0m: 0x%x (%d)\n", indentPrompt, b.PublicExponent, bigIntExponent.Int64())
	fmt.Printf("%s │ \x1b[93mModulus\x1b[0m: 0x%s\n", indentPrompt, hex.EncodeToString(b.Modulus))
	fmt.Printf("%s │ \x1b[93mAESKDFKey\x1b[0m: 0x%s\n", indentPrompt, hex.EncodeToString(b.AESKDFKey))
	fmt.Printf("%s └───\n", indentPrompt)
}

// Equal checks if two FEK_KEY_MATERIAL_BLOB structures are equal.
//
// Parameters:
// - other: The FEK_KEY_MATERIAL_BLOB structure to compare to.
//
// Returns:
// - True if the two structures are equal, false otherwise.
func (b *FEK_KEY_MATERIAL_BLOB) Equal(other *FEK_KEY_MATERIAL_BLOB) bool {
	return bytes.Equal(b.PublicExponent, other.PublicExponent) &&
		bytes.Equal(b.Modulus, other.Modulus) &&
		bytes.Equal(b.AESKDFKey, other.AESKDFKey)
}
