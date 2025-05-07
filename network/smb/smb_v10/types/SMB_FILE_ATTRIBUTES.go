package types

import "encoding/binary"

// SMB_FILE_ATTRIBUTES is a structure that contains the attributes of a file
// Sourc: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/2198f480-e047-4df0-ba64-f28eadef00b9
type SMB_FILE_ATTRIBUTES struct {
	Attributes uint16
}

// GetAttributes returns the attributes of the SMB_FILE_ATTRIBUTES structure
//
// Returns:
// - The attributes of the SMB_FILE_ATTRIBUTES structure
func (s *SMB_FILE_ATTRIBUTES) GetAttributes() uint16 {
	return s.Attributes
}

// SetAttributes sets the attributes of the SMB_FILE_ATTRIBUTES structure
//
// Parameters:
// - attributes: The attributes to set
func (s *SMB_FILE_ATTRIBUTES) SetAttributes(attributes uint16) {
	s.Attributes = attributes
}

// Marshal marshals the SMB_FILE_ATTRIBUTES structure into a byte array
//
// Returns:
// - A byte array representing the SMB_FILE_ATTRIBUTES structure
// - An error if the marshaling fails
func (s *SMB_FILE_ATTRIBUTES) Marshal() ([]byte, error) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, s.Attributes)
	return buf, nil
}

// Unmarshal unmarshals a byte array into the SMB_FILE_ATTRIBUTES structure
//
// Parameters:
// - data: The byte array to unmarshal
//
// Returns:
// - The number of bytes unmarshalled
func (s *SMB_FILE_ATTRIBUTES) Unmarshal(data []byte) (int, error) {
	s.Attributes = binary.BigEndian.Uint16(data)
	return 2, nil
}
