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
	// IoctlRequestStructureSize is the fixed StructureSize value the client MUST
	// set for an SMB2 IOCTL Request (56-byte fixed part + 1).
	IoctlRequestStructureSize = 57

	// ioctlRequestFixedSize is the size, in bytes, of the fixed portion of the body.
	ioctlRequestFixedSize = 56

	// SMB2_0_IOCTL_IS_FSCTL indicates the request is an FSCTL (rather than IOCTL).
	SMB2_0_IOCTL_IS_FSCTL = 0x00000001
)

// IoctlRequest is the SMB2 IOCTL Request body, sent by a client to issue an
// FSCTL/IOCTL control command. Output fields are 0 in a request; only the input
// buffer is carried.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/5c03c9d6-15de-48a2-9835-8fb37f8a79d8
type IoctlRequest struct {
	command_interface.Command

	// Reserved (2 bytes): The client MUST set this to 0.
	Reserved types.USHORT

	// CtlCode (4 bytes): The FSCTL/IOCTL control code.
	CtlCode types.ULONG

	// FileId (16 bytes): The file/pipe on which to perform the control.
	FileId types.SMB2_FILEID

	// MaxInputResponse (4 bytes): Max input bytes the server may return.
	MaxInputResponse types.ULONG

	// MaxOutputResponse (4 bytes): Max output bytes the server may return.
	MaxOutputResponse types.ULONG

	// Flags (4 bytes): 0 for an IOCTL, SMB2_0_IOCTL_IS_FSCTL for an FSCTL.
	Flags types.ULONG

	// Reserved2 (4 bytes): The client MUST set this to 0.
	Reserved2 types.ULONG

	// Input is the input data buffer. InputOffset/InputCount are computed on marshal.
	Input []byte
}

// NewIoctlRequest creates a new SMB2 IOCTL Request.
func NewIoctlRequest() *IoctlRequest {
	c := &IoctlRequest{Input: []byte{}}
	c.SetCommandCode(codes.SMB2_IOCTL)
	c.StructureSize = IoctlRequestStructureSize
	return c
}

// Marshal serializes the IOCTL Request body.
func (c *IoctlRequest) Marshal() ([]byte, error) {
	buf := make([]byte, ioctlRequestFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], IoctlRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	binary.LittleEndian.PutUint32(buf[4:8], c.CtlCode)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	copy(buf[8:24], fileId)

	if len(c.Input) > 0 {
		binary.LittleEndian.PutUint32(buf[24:28], uint32(header.SMB2_HEADER_SIZE+ioctlRequestFixedSize))
		binary.LittleEndian.PutUint32(buf[28:32], uint32(len(c.Input)))
	}
	binary.LittleEndian.PutUint32(buf[32:36], c.MaxInputResponse)
	// OutputOffset (36) and OutputCount (40) are 0 in a request.
	binary.LittleEndian.PutUint32(buf[44:48], c.MaxOutputResponse)
	binary.LittleEndian.PutUint32(buf[48:52], c.Flags)
	binary.LittleEndian.PutUint32(buf[52:56], c.Reserved2)

	return append(buf, c.Input...), nil
}

// Unmarshal deserializes the IOCTL Request body.
func (c *IoctlRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < ioctlRequestFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 IOCTL Request: have %d bytes, need at least %d", len(data), ioctlRequestFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	c.CtlCode = binary.LittleEndian.Uint32(data[4:8])
	if _, err := c.FileId.Unmarshal(data[8:24]); err != nil {
		return 0, err
	}
	inputOffset := int(binary.LittleEndian.Uint32(data[24:28]))
	inputCount := int(binary.LittleEndian.Uint32(data[28:32]))
	c.MaxInputResponse = binary.LittleEndian.Uint32(data[32:36])
	c.MaxOutputResponse = binary.LittleEndian.Uint32(data[44:48])
	c.Flags = binary.LittleEndian.Uint32(data[48:52])
	c.Reserved2 = binary.LittleEndian.Uint32(data[52:56])

	consumed := ioctlRequestFixedSize
	if inputCount > 0 {
		start := inputOffset - header.SMB2_HEADER_SIZE
		if start < ioctlRequestFixedSize || start+inputCount > len(data) {
			return 0, fmt.Errorf("SMB2 IOCTL Request input out of bounds: offset %d count %d", inputOffset, inputCount)
		}
		c.Input = make([]byte, inputCount)
		copy(c.Input, data[start:start+inputCount])
		consumed = start + inputCount
	} else {
		c.Input = []byte{}
	}

	return consumed, nil
}
