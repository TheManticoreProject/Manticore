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
	// ReadRequestStructureSize is the fixed StructureSize value the client MUST
	// set for an SMB2 READ Request (48-byte fixed part + 1).
	ReadRequestStructureSize = 49

	// readRequestFixedSize is the size, in bytes, of the fixed portion of the body.
	readRequestFixedSize = 48
)

// ReadRequest is the SMB2 READ Request body, sent by the client to read from the
// file or pipe identified by FileId.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/320f04f3-1b28-45cd-aaa1-9e5aed810dca
type ReadRequest struct {
	command_interface.Command

	// Padding (1 byte): Requested placement offset of the data in the response.
	Padding types.UCHAR

	// Flags (1 byte): Read flags (SMB 3.x only; else 0).
	Flags types.UCHAR

	// Length (4 bytes): The number of bytes to read.
	Length types.ULONG

	// Offset (8 bytes): The byte offset into the file to read from.
	Offset types.UINT64

	// FileId (16 bytes): The file or pipe to read from.
	FileId types.SMB2_FILEID

	// MinimumCount (4 bytes): The minimum number of bytes for the read to succeed.
	MinimumCount types.ULONG

	// Channel (4 bytes): RDMA channel selector (SMB 3.x only; else 0).
	Channel types.ULONG

	// RemainingBytes (4 bytes): RDMA read length (SMB 3.x only; else 0).
	RemainingBytes types.ULONG

	// ReadChannelInfo is the raw RDMA channel-info buffer.
	// ReadChannelInfoOffset/Length are computed on marshal.
	ReadChannelInfo []byte
}

// NewReadRequest creates a new SMB2 READ Request.
func NewReadRequest() *ReadRequest {
	c := &ReadRequest{ReadChannelInfo: []byte{}}
	c.SetCommandCode(codes.SMB2_READ)
	c.StructureSize = ReadRequestStructureSize
	return c
}

// Marshal serializes the READ Request body.
func (c *ReadRequest) Marshal() ([]byte, error) {
	buf := make([]byte, readRequestFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], ReadRequestStructureSize)
	buf[2] = byte(c.Padding)
	buf[3] = byte(c.Flags)
	binary.LittleEndian.PutUint32(buf[4:8], c.Length)
	binary.LittleEndian.PutUint64(buf[8:16], c.Offset)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	copy(buf[16:32], fileId)

	binary.LittleEndian.PutUint32(buf[32:36], c.MinimumCount)
	binary.LittleEndian.PutUint32(buf[36:40], c.Channel)
	binary.LittleEndian.PutUint32(buf[40:44], c.RemainingBytes)

	if len(c.ReadChannelInfo) > 0 {
		binary.LittleEndian.PutUint16(buf[44:46], uint16(header.SMB2_HEADER_SIZE+readRequestFixedSize))
		binary.LittleEndian.PutUint16(buf[46:48], uint16(len(c.ReadChannelInfo)))
		buf = append(buf, c.ReadChannelInfo...)
	} else {
		// The Buffer field must be at least one byte in length.
		buf = append(buf, 0x00)
	}

	return buf, nil
}

// Unmarshal deserializes the READ Request body.
func (c *ReadRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < readRequestFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 READ Request: have %d bytes, need at least %d", len(data), readRequestFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Padding = data[2]
	c.Flags = data[3]
	c.Length = binary.LittleEndian.Uint32(data[4:8])
	c.Offset = binary.LittleEndian.Uint64(data[8:16])
	if _, err := c.FileId.Unmarshal(data[16:32]); err != nil {
		return 0, err
	}
	c.MinimumCount = binary.LittleEndian.Uint32(data[32:36])
	c.Channel = binary.LittleEndian.Uint32(data[36:40])
	c.RemainingBytes = binary.LittleEndian.Uint32(data[40:44])
	channelInfoOffset := int(binary.LittleEndian.Uint16(data[44:46]))
	channelInfoLength := int(binary.LittleEndian.Uint16(data[46:48]))

	consumed := readRequestFixedSize
	if channelInfoLength > 0 {
		start := channelInfoOffset - header.SMB2_HEADER_SIZE
		if start < readRequestFixedSize || start+channelInfoLength > len(data) {
			return 0, fmt.Errorf("SMB2 READ Request channel info out of bounds: offset %d length %d", channelInfoOffset, channelInfoLength)
		}
		c.ReadChannelInfo = make([]byte, channelInfoLength)
		copy(c.ReadChannelInfo, data[start:start+channelInfoLength])
		consumed = start + channelInfoLength
	} else {
		c.ReadChannelInfo = []byte{}
	}

	return consumed, nil
}
