package key

import (
	"fmt"
	"strings"
)

// CustomKeyInformationVolumeType represents the volume type.
//
// See: https://msdn.microsoft.com/en-us/library/mt220496.aspx
type CustomKeyInformationVolumeType struct {
	Value uint8

	// Internal
	RawBytes     []byte
	RawBytesSize uint32
}

const (
	// Volume not specified.
	CustomKeyInformationVolumeType_None uint8 = 0x00

	// Operating system volume (OSV).
	CustomKeyInformationVolumeType_OperatingSystem uint8 = 0x01

	// Fixed data volume (FDV).
	CustomKeyInformationVolumeType_Fixed uint8 = 0x02

	// Removable data volume (RDV).
	CustomKeyInformationVolumeType_Removable uint8 = 0x03
)

// Parse parses the provided byte slice into the CustomKeyInformationVolumeType structure.
//
// Parameters:
// - value: A byte slice containing the raw volume type to be parsed.
//
// Note:
// The function expects the byte slice to contain a single byte representing the volume type.
// It extracts the volume type value from the byte slice and assigns it to the CustomKeyInformationVolumeType structure.
func (vt *CustomKeyInformationVolumeType) FromBytes(value byte) {
	vt.Value = value
}

// String returns a string representation of the CustomKeyInformationVolumeType.
//
// Returns:
// - A string representing the CustomKeyInformationVolumeType.
func (vt *CustomKeyInformationVolumeType) String() string {
	switch vt.Value {
	case CustomKeyInformationVolumeType_OperatingSystem:
		return "Operating System"
	case CustomKeyInformationVolumeType_Fixed:
		return "Fixed"
	case CustomKeyInformationVolumeType_Removable:
		return "Removable"
	default:
		return "None"
	}
}

// Describe prints a detailed description of the CustomKeyInformationVolumeType instance.
//
// Parameters:
// - indent: An integer representing the indentation level for the printed output.
//
// Note:
// This function prints the Value and Name of the CustomKeyInformationVolumeType instance.
// The output is formatted with the specified indentation level to improve readability.
func (vt *CustomKeyInformationVolumeType) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mCustomKeyInformationVolumeType\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mValue\x1b[0m: %d\n", indentPrompt, vt.Value)
	fmt.Printf("%s │ \x1b[93mName\x1b[0m: %s\n", indentPrompt, vt.String())
	fmt.Printf("%s └─\n", indentPrompt)
}
