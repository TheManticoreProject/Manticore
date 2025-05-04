package types

type OEM_STRING struct {
	Length uint16
	Data   []byte
}

// Marshal marshals the OEM_STRING into a byte array
//
// Returns:
// - The byte array representing the OEM_STRING
func (o *OEM_STRING) Marshal() ([]byte, error) {
	return o.Data, nil
}

// Unmarshal unmarshals the OEM_STRING from a byte array
//
// Parameters:
// - data: The byte array to unmarshal the OEM_STRING from
//
// Returns:
// - The number of bytes unmarshalled
func (o *OEM_STRING) Unmarshal(data []byte) error {
	return nil
}
