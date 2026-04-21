package headers

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// FEK_KEY_MATERIAL_HEADER describes the sizes of the two components that make up
// the KEY_USAGE_FEK key material buffer: the RSA 2048 public key (PublicExponent
// and Modulus) and the AES-256 KDF key.
//
// The header immediately follows the 1-byte FEK_KEY_MATERIAL_VERSION and precedes
// the raw RSA and AES-256 KDF key bytes. All integer fields are stored in little-endian.
//
// See:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/735fd27a-3f22-4926-93f9-0298bb67a84b
type FEK_KEY_MATERIAL_HEADER struct {
	// BitLength is the size, in bits, of the RSA key. For KEY_USAGE_FEK this MUST be 2048.
	BitLength uint32

	// CbPublicExp is the size, in bytes, of the RSA public exponent.
	CbPublicExp uint32

	// CbModulus is the size, in bytes, of the RSA modulus.
	CbModulus uint32

	// CbAESKDFKey is the size, in bytes, of the AES-256 KDF key. For KEY_USAGE_FEK this MUST be 32.
	CbAESKDFKey uint32
}

// Unmarshal parses the provided byte slice into the FEK_KEY_MATERIAL_HEADER structure.
//
// Parameters:
// - value: A byte slice containing the raw FEK key material header to be parsed.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
func (h *FEK_KEY_MATERIAL_HEADER) Unmarshal(value []byte) (int, error) {
	if len(value) < 16 {
		return 0, errors.New("buffer too small for FEK_KEY_MATERIAL_HEADER, at least 16 bytes are required")
	}

	bytesRead := 0

	h.BitLength = binary.LittleEndian.Uint32(value[bytesRead : bytesRead+4])
	bytesRead += 4

	h.CbPublicExp = binary.LittleEndian.Uint32(value[bytesRead : bytesRead+4])
	bytesRead += 4

	h.CbModulus = binary.LittleEndian.Uint32(value[bytesRead : bytesRead+4])
	bytesRead += 4

	h.CbAESKDFKey = binary.LittleEndian.Uint32(value[bytesRead : bytesRead+4])
	bytesRead += 4

	return bytesRead, nil
}

// Marshal returns the raw bytes of the FEK_KEY_MATERIAL_HEADER structure.
//
// Returns:
// - A byte slice representing the raw bytes of the FEK_KEY_MATERIAL_HEADER structure.
func (h *FEK_KEY_MATERIAL_HEADER) Marshal() ([]byte, error) {
	buf := make([]byte, 16)
	bytesWritten := 0

	binary.LittleEndian.PutUint32(buf[bytesWritten:bytesWritten+4], h.BitLength)
	bytesWritten += 4

	binary.LittleEndian.PutUint32(buf[bytesWritten:bytesWritten+4], h.CbPublicExp)
	bytesWritten += 4

	binary.LittleEndian.PutUint32(buf[bytesWritten:bytesWritten+4], h.CbModulus)
	bytesWritten += 4

	binary.LittleEndian.PutUint32(buf[bytesWritten:bytesWritten+4], h.CbAESKDFKey)
	bytesWritten += 4

	return buf, nil
}

// Describe prints a detailed description of the FEK_KEY_MATERIAL_HEADER instance.
//
// Parameters:
// - indent: An integer representing the indentation level for the printed output.
func (h *FEK_KEY_MATERIAL_HEADER) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mFEK_KEY_MATERIAL_HEADER (header)\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mBitLength\x1b[0m   : (0x%08x) %d bits \n", indentPrompt, h.BitLength, h.BitLength)
	fmt.Printf("%s │ \x1b[93mCbPublicExp\x1b[0m : (0x%08x) %d bytes \n", indentPrompt, h.CbPublicExp, h.CbPublicExp)
	fmt.Printf("%s │ \x1b[93mCbModulus\x1b[0m   : (0x%08x) %d bytes \n", indentPrompt, h.CbModulus, h.CbModulus)
	fmt.Printf("%s │ \x1b[93mCbAESKDFKey\x1b[0m : (0x%08x) %d bytes \n", indentPrompt, h.CbAESKDFKey, h.CbAESKDFKey)
	fmt.Printf("%s └───\n", indentPrompt)
}

// Equal checks if two FEK_KEY_MATERIAL_HEADER structures are equal.
//
// Parameters:
// - other: The FEK_KEY_MATERIAL_HEADER structure to compare to.
//
// Returns:
// - True if the two structures are equal, false otherwise.
func (h *FEK_KEY_MATERIAL_HEADER) Equal(other *FEK_KEY_MATERIAL_HEADER) bool {
	return h.BitLength == other.BitLength &&
		h.CbPublicExp == other.CbPublicExp &&
		h.CbModulus == other.CbModulus &&
		h.CbAESKDFKey == other.CbAESKDFKey
}

// String returns a string representation of the FEK_KEY_MATERIAL_HEADER structure.
//
// Returns:
// - A string representing the FEK_KEY_MATERIAL_HEADER structure.
func (h *FEK_KEY_MATERIAL_HEADER) String() string {
	return fmt.Sprintf(
		"FEK_KEY_MATERIAL_HEADER(BitLength: %d, CbPublicExp: %d, CbModulus: %d, CbAESKDFKey: %d)",
		h.BitLength, h.CbPublicExp, h.CbModulus, h.CbAESKDFKey,
	)
}
