package andx

import (
	"encoding/binary"
	"errors"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
)

// AndX Messages contain a construct, conceptually similar to a linked-list, that is used to connect the batched block pairs.
// Source: 2.2.3.4 Batched Messages ("AndX" Messages) https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/fc4d19f7-8040-426d-9154-7219c57453c8
type AndX struct {
	AndXCommand  codes.CommandCode
	AndXReserved uint8
	AndXOffset   uint16
}

// NewAndX creates a new AndX structure
func NewAndX() *AndX {
	return &AndX{
		AndXCommand:  0,
		AndXReserved: 0,
		AndXOffset:   0,
	}
}

// GetParameters returns the parameters of the AndX structure
// Returns:
// - A byte array containing the parameters of the AndX structure
func (a *AndX) GetParameters() []uint16 {
	return []uint16{uint16(a.AndXCommand)<<8 | uint16(a.AndXReserved), a.AndXOffset}
}

// Marshal marshals the AndX structure into a byte array
// Returns:
// - A byte array containing the marshalled AndX structure
// - An error if the marshalling process fails, or nil if successful
func (a *AndX) Marshal() ([]byte, error) {
	marshalled_andx := []byte{}

	marshalled_andx = append(marshalled_andx, byte(a.AndXCommand))

	marshalled_andx = append(marshalled_andx, a.AndXReserved)

	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, a.AndXOffset)
	marshalled_andx = append(marshalled_andx, buf2...)

	return marshalled_andx, nil
}

// Unmarshal unmarshals the AndX structure from a byte array
// Returns:
// - The number of bytes read
// - An error if the unmarshalling process fails, or nil if successful
func (a *AndX) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, errors.New("data is too short to unmarshal AndX")
	}

	a.AndXCommand = codes.CommandCode(data[0])

	a.AndXReserved = data[1]

	a.AndXOffset = binary.BigEndian.Uint16(data[2:4])

	return 4, nil
}

// GetCommandCode returns the command code of the AndX structure
// Returns:
// - The command code of the AndX structure
func (a *AndX) GetCommandCode() codes.CommandCode {
	return a.AndXCommand
}

// GetOffset returns the offset of the AndX structure
// Returns:
// - The offset of the AndX structure
func (a *AndX) GetOffset() uint16 {
	return a.AndXOffset
}
