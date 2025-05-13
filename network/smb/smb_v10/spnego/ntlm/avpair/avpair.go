package avpair

import (
	"encoding/binary"
	"fmt"
)

// AV_PAIR
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/83f5e789-660d-4781-8491-5f8c6641f75e
type AvPair struct {
	// AvId (2 bytes): A 16-bit unsigned integer that defines the
	// information type in the Value field. The contents of this field
	// MUST be a value from the following table. The corresponding Value
	// field in this AV_PAIR MUST contain the information specified in the
	// description of that AvId.
	AvID uint16

	// AvLen (2 bytes): A 16-bit unsigned integer that defines the length,
	// in bytes, of the Value field.
	AvLen uint16

	// Value (variable): A variable-length byte-array that contains the value
	// defined for this AV pair entry. The contents of this field depend on
	// the type expressed in the AvId field. The available types and resulting
	// format and contents of this field are specified in the table within
	// the AvId field description in this topic.
	AvData []byte
}

// String returns a string representation of the AV_PAIR.
func (a *AvPair) String() string {
	return fmt.Sprintf("AvId: %d, AvLen: %d, AvData: %v", a.AvID, a.AvLen, a.AvData)
}

// Marshal serializes the AV_PAIR to a byte slice.
func (a *AvPair) Marshal() ([]byte, error) {
	marshaledData := []byte{}

	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, a.AvID)
	marshaledData = append(marshaledData, buf2...)

	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, a.AvLen)
	marshaledData = append(marshaledData, buf2...)

	marshaledData = append(marshaledData, a.AvData...)

	return marshaledData, nil
}

// Unmarshal parses the AV_PAIR from a byte slice.
func (a *AvPair) Unmarshal(marshaledData []byte) (int, error) {
	buf := []byte{}

	a.AvID = binary.LittleEndian.Uint16(marshaledData[0:2])

	a.AvLen = binary.LittleEndian.Uint16(marshaledData[2:4])

	a.AvData = marshaledData[4:]

	return len(buf), nil
}
