package magic

import (
	"errors"
	"fmt"
	"strings"
)

// FEK_KEY_MATERIAL_VERSION is the 1-byte version tag that prefixes the FEK key
// material buffer described in MS-ADTS 2.2.20.5.3.
//
// See:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/735fd27a-3f22-4926-93f9-0298bb67a84b
type FEK_KEY_MATERIAL_VERSION struct {
	// Version identifies the layout of the FEK key material buffer.
	// This field MUST be set to FEK_KEY_VERSION_1 (0x01).
	Version uint8
}

// Unmarshal parses the provided byte slice into the FEK_KEY_MATERIAL_VERSION structure.
//
// Parameters:
// - value: A byte slice containing the raw version byte to be parsed.
//
// Returns:
// - The number of bytes read from the byte slice.
// - An error if the parsing fails, otherwise nil.
func (v *FEK_KEY_MATERIAL_VERSION) Unmarshal(value []byte) (int, error) {
	if len(value) < 1 {
		return 0, errors.New("buffer too small for FEK_KEY_MATERIAL_VERSION, at least 1 byte is required")
	}

	v.Version = value[0]

	return 1, nil
}

// Marshal returns the raw bytes of the FEK_KEY_MATERIAL_VERSION structure.
//
// Returns:
// - A byte slice representing the raw bytes of the FEK_KEY_MATERIAL_VERSION structure.
// - An error if the conversion fails.
func (v *FEK_KEY_MATERIAL_VERSION) Marshal() ([]byte, error) {
	return []byte{v.Version}, nil
}

// String returns a string representation of the FEK_KEY_MATERIAL_VERSION structure.
//
// Returns:
// - A string representing the FEK_KEY_MATERIAL_VERSION structure.
func (v *FEK_KEY_MATERIAL_VERSION) String() string {
	return fmt.Sprintf("FEK_KEY_MATERIAL_VERSION: Version=0x%02x", v.Version)
}

// Equal returns true if the FEK_KEY_MATERIAL_VERSION structure is equal to the other FEK_KEY_MATERIAL_VERSION structure.
//
// Parameters:
// - other: The other FEK_KEY_MATERIAL_VERSION structure to compare to.
//
// Returns:
// - True if the structures are equal, otherwise false.
func (v *FEK_KEY_MATERIAL_VERSION) Equal(other *FEK_KEY_MATERIAL_VERSION) bool {
	return v.Version == other.Version
}

// Describe prints the FEK_KEY_MATERIAL_VERSION structure to the console.
//
// Parameters:
// - indent: The number of levels to indent the output.
func (v *FEK_KEY_MATERIAL_VERSION) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mFEK_KEY_MATERIAL_VERSION (magic)\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mVersion\x1b[0m: 0x%02x\n", indentPrompt, v.Version)
	fmt.Printf("%s └───\n", indentPrompt)
}
