package customkeyinformation

import (
	"encoding/hex"
	"fmt"
)

// Reserved represents the Reserved structure.
//
// See: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/701a55dc-d062-4032-a2da-dbdfc384c8cf
type Reserved []byte

// reservedMaxSize is the size [MS-ADTS] gives the Reserved field. Real keys carry
// a shorter one (an Azure AD FIDO2 key ends nine bytes into it), so it is read up
// to this size and re-emitted at the length it was read.
const reservedMaxSize = 10

// Unmarshal reads the Reserved field: up to 10 bytes, fewer when the structure
// ends inside it.
//
// Parameters:
// - data: A byte slice containing the raw reserved to be parsed.
//
// Returns:
// - The number of bytes read from the data.
// - An error if the data is empty.
func (r *Reserved) Unmarshal(data []byte) (int, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("invalid data length: %d", len(data))
	}

	n := min(len(data), reservedMaxSize)
	*r = append(Reserved(nil), data[:n]...)

	return n, nil
}

// Marshal returns the raw bytes of the Reserved structure, at the length it was
// read or set.
//
// Returns:
// - A byte slice representing the raw bytes of the Reserved structure.
// - An error if the field is longer than 10 bytes.
func (r *Reserved) Marshal() ([]byte, error) {
	if len(*r) > reservedMaxSize {
		return nil, fmt.Errorf("Reserved is %d bytes, longer than the %d-byte field", len(*r), reservedMaxSize)
	}
	return append([]byte(nil), (*r)...), nil
}

// String returns a string representation of the Reserved.
//
// Returns:
// - A string representing the Reserved.
func (r *Reserved) String() string {
	return hex.EncodeToString(*r)
}
