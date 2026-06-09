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
	// IoctlResponseStructureSize is the fixed StructureSize value the server MUST
	// set for an SMB2 IOCTL Response (48-byte fixed part + 1).
	IoctlResponseStructureSize = 49

	// ioctlResponseFixedSize is the size, in bytes, of the fixed portion of the body.
	ioctlResponseFixedSize = 48
)

// IoctlResponse is the SMB2 IOCTL Response body, sent by the server with the
// result of an FSCTL/IOCTL command.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/f70eccb6-e1be-4db8-9c47-9ac86ef18dbb
type IoctlResponse struct {
	command_interface.Command

	// Reserved (2 bytes): The server MUST set this to 0.
	Reserved types.USHORT

	// CtlCode (4 bytes): The FSCTL/IOCTL control code that was executed.
	CtlCode types.ULONG

	// FileId (16 bytes): The file/pipe on which the control was performed.
	FileId types.SMB2_FILEID

	// Flags (4 bytes): Reserved; the server MUST set this to 0.
	Flags types.ULONG

	// Reserved2 (4 bytes): The server MUST set this to 0.
	Reserved2 types.ULONG

	// Input is the returned input data buffer (often empty).
	Input []byte

	// Output is the returned output data buffer.
	Output []byte
}

// NewIoctlResponse creates a new SMB2 IOCTL Response.
func NewIoctlResponse() *IoctlResponse {
	c := &IoctlResponse{Input: []byte{}, Output: []byte{}}
	c.SetCommandCode(codes.SMB2_IOCTL)
	c.StructureSize = IoctlResponseStructureSize
	return c
}

// Marshal serializes the IOCTL Response body. The output buffer, if present,
// begins at the first 8-byte-aligned offset after the input buffer.
func (c *IoctlResponse) Marshal() ([]byte, error) {
	buf := make([]byte, ioctlResponseFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], IoctlResponseStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	binary.LittleEndian.PutUint32(buf[4:8], c.CtlCode)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	copy(buf[8:24], fileId)

	variable := []byte{}
	if len(c.Input) > 0 {
		binary.LittleEndian.PutUint32(buf[24:28], uint32(header.SMB2_HEADER_SIZE+ioctlResponseFixedSize))
		binary.LittleEndian.PutUint32(buf[28:32], uint32(len(c.Input)))
		variable = append(variable, c.Input...)
	}
	if len(c.Output) > 0 {
		pos := header.SMB2_HEADER_SIZE + ioctlResponseFixedSize + len(variable)
		pad := align8(pos) - pos
		variable = append(variable, make([]byte, pad)...)
		binary.LittleEndian.PutUint32(buf[32:36], uint32(header.SMB2_HEADER_SIZE+ioctlResponseFixedSize+len(variable)))
		binary.LittleEndian.PutUint32(buf[36:40], uint32(len(c.Output)))
		variable = append(variable, c.Output...)
	}
	binary.LittleEndian.PutUint32(buf[40:44], c.Flags)
	binary.LittleEndian.PutUint32(buf[44:48], c.Reserved2)

	return append(buf, variable...), nil
}

// Unmarshal deserializes the IOCTL Response body.
func (c *IoctlResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < ioctlResponseFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 IOCTL Response: have %d bytes, need at least %d", len(data), ioctlResponseFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	c.CtlCode = binary.LittleEndian.Uint32(data[4:8])
	if _, err := c.FileId.Unmarshal(data[8:24]); err != nil {
		return 0, err
	}
	inputOffset := int(binary.LittleEndian.Uint32(data[24:28]))
	inputCount := int(binary.LittleEndian.Uint32(data[28:32]))
	outputOffset := int(binary.LittleEndian.Uint32(data[32:36]))
	outputCount := int(binary.LittleEndian.Uint32(data[36:40]))
	c.Flags = binary.LittleEndian.Uint32(data[40:44])
	c.Reserved2 = binary.LittleEndian.Uint32(data[44:48])

	consumed := ioctlResponseFixedSize
	if inputCount > 0 {
		start := inputOffset - header.SMB2_HEADER_SIZE
		if start < ioctlResponseFixedSize || start+inputCount > len(data) {
			return 0, fmt.Errorf("SMB2 IOCTL Response input out of bounds: offset %d count %d", inputOffset, inputCount)
		}
		c.Input = make([]byte, inputCount)
		copy(c.Input, data[start:start+inputCount])
		if start+inputCount > consumed {
			consumed = start + inputCount
		}
	} else {
		c.Input = []byte{}
	}

	if outputCount > 0 {
		start := outputOffset - header.SMB2_HEADER_SIZE
		if start < ioctlResponseFixedSize || start+outputCount > len(data) {
			return 0, fmt.Errorf("SMB2 IOCTL Response output out of bounds: offset %d count %d", outputOffset, outputCount)
		}
		c.Output = make([]byte, outputCount)
		copy(c.Output, data[start:start+outputCount])
		if start+outputCount > consumed {
			consumed = start + outputCount
		}
	} else {
		c.Output = []byte{}
	}

	return consumed, nil
}
