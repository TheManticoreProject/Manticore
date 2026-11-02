package customkeyinformation

import "fmt"

// SupportsNotification is the 1-byte SupportsNotification field of a
// CUSTOM_KEY_INFORMATION structure ([MS-ADTS] 2.2.20.2.1). Zero means the device
// does not support notification and any other value that it does. The byte is
// kept as read, not reduced to a bool, so a value other than 0 or 1 is written
// back unchanged and the entry (covered by the KeyHash) round-trips.
type SupportsNotification uint8

const (
	SupportsNotification_TRUE  SupportsNotification = 1
	SupportsNotification_FALSE SupportsNotification = 0
)

// Bool reports whether the field indicates notification support (non-zero).
func (sn SupportsNotification) Bool() bool {
	return sn != 0
}

// Marshal returns the raw byte of the SupportsNotification field.
//
// Returns:
// - A byte slice holding the field's byte.
// - An error if the conversion fails.
func (sn SupportsNotification) Marshal() ([]byte, error) {
	return []byte{byte(sn)}, nil
}

// Unmarshal parses the provided byte slice into the SupportsNotification structure.
//
// Parameters:
// - data: A byte slice containing the raw supports notification to be parsed.
//
// Returns:
// - The number of bytes read from the data.
// - An error if the parsing fails, otherwise nil.
func (sn *SupportsNotification) Unmarshal(data []byte) (int, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("invalid data length: %d", len(data))
	}

	*sn = SupportsNotification(data[0])

	return 1, nil
}

// String returns a string representation of the SupportsNotification: "True" or
// "False", with the raw value appended when it is neither 0 nor 1.
//
// Returns:
// - A string representing the SupportsNotification.
func (sn SupportsNotification) String() string {
	switch sn {
	case SupportsNotification_FALSE:
		return "False"
	case SupportsNotification_TRUE:
		return "True"
	default:
		return fmt.Sprintf("True (0x%02x)", uint8(sn))
	}
}
