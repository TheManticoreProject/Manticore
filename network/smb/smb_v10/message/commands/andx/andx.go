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

// GetParameters returns the AndX block as the words a Parameters structure
// carries it in.
//
// Parameters is a byte-transparent container: it splits every word high byte
// first and reassembles it the same way, so a word holds two consecutive wire
// bytes rather than a 16-bit value. AndXCommand and AndXReserved are single
// bytes and pack into the first word directly. AndXOffset is a little-endian
// USHORT ([MS-CIFS] 2.2.3.4), so its two bytes are packed in wire order here —
// returning the field as a value put it on the wire with its bytes reversed.
//
// Returns:
// - The words of the AndX block, in wire order
func (a *AndX) GetParameters() []uint16 {
	return []uint16{
		uint16(a.AndXCommand)<<8 | uint16(a.AndXReserved),
		uint16(a.AndXOffset&0x00FF)<<8 | uint16(a.AndXOffset>>8),
	}
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
	binary.LittleEndian.PutUint16(buf2, a.AndXOffset)
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

	a.AndXOffset = binary.LittleEndian.Uint16(data[2:4])

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
