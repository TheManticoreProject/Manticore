package strength

import (
	"fmt"
)

// KeyStrength specifies the strength of the NGC key.
// See: https://msdn.microsoft.com/en-us/library/mt220496.aspx
type KeyStrength uint32

const (
	// Key strength is unknown.
	KeyStrength_Unknown KeyStrength = 0x00

	// Key strength is weak.
	KeyStrength_Weak KeyStrength = 0x01

	// Key strength is normal.
	KeyStrength_Normal KeyStrength = 0x02
)

// Unmarshal parses the 1-byte KeyStrength field of a CUSTOM_KEY_INFORMATION
// structure ([MS-ADTS] 2.2.20.2.1).
//
// Returns:
// - The number of bytes read from the data.
// - An error if the data is empty.
func (ks *KeyStrength) Unmarshal(data []byte) (int, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("invalid data length: %d", len(data))
	}

	*ks = KeyStrength(data[0])

	return 1, nil
}

// Marshal returns the 1-byte wire form of the KeyStrength field.
//
// Returns:
// - A byte slice representing the raw bytes of the KeyStrength structure.
// - An error if the value does not fit in one byte.
func (ks *KeyStrength) Marshal() ([]byte, error) {
	if *ks > 0xFF {
		return nil, fmt.Errorf("KeyStrength 0x%x does not fit the 1-byte field", uint32(*ks))
	}
	return []byte{byte(*ks)}, nil
}

func (ks KeyStrength) String() string {
	switch ks {
	case KeyStrength_Unknown:
		return "Unknown"
	case KeyStrength_Weak:
		return "Weak"
	case KeyStrength_Normal:
		return "Normal"
	}
	return "Unknown"
}
