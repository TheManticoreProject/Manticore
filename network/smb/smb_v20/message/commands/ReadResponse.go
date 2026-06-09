package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

const (
	// ReadResponseStructureSize is the fixed StructureSize value the server MUST
	// set for an SMB2 READ Response (16-byte fixed part + 1).
	ReadResponseStructureSize = 17

	// readResponseFixedSize is the size, in bytes, of the fixed portion of the body.
	readResponseFixedSize = 16
)

// ReadResponse is the SMB2 READ Response body, sent by the server with the data
// read from a file or pipe.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/3e3d2f2c-0e2f-41ea-ad07-fbca6ffdfd90
type ReadResponse struct {
	command_interface.Command

	// Reserved (1 byte): The server MUST set this to 0.
	Reserved types.UCHAR

	// DataRemaining (4 bytes): The number of bytes remaining to be sent on the channel.
	DataRemaining types.ULONG

	// Flags (4 bytes): Reserved2 for SMB < 3.1.1; read-response flags for SMB 3.1.1.
	Flags types.ULONG

	// Data is the data read. DataOffset/DataLength are computed on marshal.
	Data []byte
}

// NewReadResponse creates a new SMB2 READ Response.
func NewReadResponse() *ReadResponse {
	c := &ReadResponse{Data: []byte{}}
	c.SetCommandCode(codes.SMB2_READ)
	c.StructureSize = ReadResponseStructureSize
	return c
}

// Marshal serializes the READ Response body.
func (c *ReadResponse) Marshal() ([]byte, error) {
	buf := make([]byte, readResponseFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], ReadResponseStructureSize)
	if len(c.Data) > 0 {
		buf[2] = byte(header.SMB2_HEADER_SIZE + readResponseFixedSize) // DataOffset
	}
	buf[3] = byte(c.Reserved)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(c.Data)))
	binary.LittleEndian.PutUint32(buf[8:12], c.DataRemaining)
	binary.LittleEndian.PutUint32(buf[12:16], c.Flags)

	if len(c.Data) > 0 {
		buf = append(buf, c.Data...)
	} else {
		// The Buffer field has a minimum length of 1 byte.
		buf = append(buf, 0x00)
	}

	return buf, nil
}

// Unmarshal deserializes the READ Response body.
func (c *ReadResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < readResponseFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 READ Response: have %d bytes, need at least %d", len(data), readResponseFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	dataOffset := int(data[2])
	c.Reserved = data[3]
	dataLength := int(binary.LittleEndian.Uint32(data[4:8]))
	c.DataRemaining = binary.LittleEndian.Uint32(data[8:12])
	c.Flags = binary.LittleEndian.Uint32(data[12:16])

	consumed := readResponseFixedSize
	if dataLength > 0 {
		start := dataOffset - header.SMB2_HEADER_SIZE
		if start < readResponseFixedSize || start+dataLength > len(data) {
			return 0, fmt.Errorf("SMB2 READ Response data out of bounds: offset %d length %d", dataOffset, dataLength)
		}
		c.Data = make([]byte, dataLength)
		copy(c.Data, data[start:start+dataLength])
		consumed = start + dataLength
	} else {
		c.Data = []byte{}
	}

	return consumed, nil
}
