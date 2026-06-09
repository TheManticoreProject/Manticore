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
	// WriteRequestStructureSize is the fixed StructureSize value the client MUST
	// set for an SMB2 WRITE Request (48-byte fixed part + 1).
	WriteRequestStructureSize = 49

	// writeRequestFixedSize is the size, in bytes, of the fixed portion of the body.
	writeRequestFixedSize = 48
)

// WriteRequest is the SMB2 WRITE Request body, sent by the client to write data
// to the file or pipe identified by FileId.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/e7046961-3318-4350-be2a-a8d69bb59ce8
type WriteRequest struct {
	command_interface.Command

	// Offset (8 bytes): The byte offset in the destination file to write at.
	Offset types.UINT64

	// FileId (16 bytes): The file or pipe to write to.
	FileId types.SMB2_FILEID

	// Channel (4 bytes): RDMA channel selector (SMB 3.x only; else 0).
	Channel types.ULONG

	// RemainingBytes (4 bytes): RDMA write length (SMB 3.x only; else 0).
	RemainingBytes types.ULONG

	// Flags (4 bytes): Write flags (write-through / unbuffered; SMB > 2.0.2).
	Flags types.ULONG

	// Data is the data to write. DataOffset/Length are computed on marshal.
	Data []byte

	// WriteChannelInfo is the raw RDMA channel-info buffer (placed after Data).
	WriteChannelInfo []byte
}

// NewWriteRequest creates a new SMB2 WRITE Request.
func NewWriteRequest() *WriteRequest {
	c := &WriteRequest{Data: []byte{}, WriteChannelInfo: []byte{}}
	c.SetCommandCode(codes.SMB2_WRITE)
	c.StructureSize = WriteRequestStructureSize
	return c
}

// Marshal serializes the WRITE Request body.
func (c *WriteRequest) Marshal() ([]byte, error) {
	buf := make([]byte, writeRequestFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], WriteRequestStructureSize)

	// Data begins immediately after the fixed part; offsets are header-relative.
	if len(c.Data) > 0 {
		binary.LittleEndian.PutUint16(buf[2:4], uint16(header.SMB2_HEADER_SIZE+writeRequestFixedSize))
	}
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(c.Data)))
	binary.LittleEndian.PutUint64(buf[8:16], c.Offset)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	copy(buf[16:32], fileId)

	binary.LittleEndian.PutUint32(buf[32:36], c.Channel)
	binary.LittleEndian.PutUint32(buf[36:40], c.RemainingBytes)

	variable := append([]byte{}, c.Data...)
	if len(c.WriteChannelInfo) > 0 {
		binary.LittleEndian.PutUint16(buf[40:42], uint16(header.SMB2_HEADER_SIZE+writeRequestFixedSize+len(variable)))
		binary.LittleEndian.PutUint16(buf[42:44], uint16(len(c.WriteChannelInfo)))
		variable = append(variable, c.WriteChannelInfo...)
	}
	binary.LittleEndian.PutUint32(buf[44:48], c.Flags)

	if len(variable) == 0 {
		// Keep a non-empty Buffer to match the StructureSize off-by-one convention.
		variable = []byte{0x00}
	}

	return append(buf, variable...), nil
}

// Unmarshal deserializes the WRITE Request body.
func (c *WriteRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < writeRequestFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 WRITE Request: have %d bytes, need at least %d", len(data), writeRequestFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	dataOffset := int(binary.LittleEndian.Uint16(data[2:4]))
	dataLength := int(binary.LittleEndian.Uint32(data[4:8]))
	c.Offset = binary.LittleEndian.Uint64(data[8:16])
	if _, err := c.FileId.Unmarshal(data[16:32]); err != nil {
		return 0, err
	}
	c.Channel = binary.LittleEndian.Uint32(data[32:36])
	c.RemainingBytes = binary.LittleEndian.Uint32(data[36:40])
	channelInfoOffset := int(binary.LittleEndian.Uint16(data[40:42]))
	channelInfoLength := int(binary.LittleEndian.Uint16(data[42:44]))
	c.Flags = binary.LittleEndian.Uint32(data[44:48])

	consumed := writeRequestFixedSize
	if dataLength > 0 {
		start := dataOffset - header.SMB2_HEADER_SIZE
		if start < writeRequestFixedSize || start+dataLength > len(data) {
			return 0, fmt.Errorf("SMB2 WRITE Request data out of bounds: offset %d length %d", dataOffset, dataLength)
		}
		c.Data = make([]byte, dataLength)
		copy(c.Data, data[start:start+dataLength])
		if start+dataLength > consumed {
			consumed = start + dataLength
		}
	} else {
		c.Data = []byte{}
	}

	if channelInfoLength > 0 {
		start := channelInfoOffset - header.SMB2_HEADER_SIZE
		if start < writeRequestFixedSize || start+channelInfoLength > len(data) {
			return 0, fmt.Errorf("SMB2 WRITE Request channel info out of bounds: offset %d length %d", channelInfoOffset, channelInfoLength)
		}
		c.WriteChannelInfo = make([]byte, channelInfoLength)
		copy(c.WriteChannelInfo, data[start:start+channelInfoLength])
		if start+channelInfoLength > consumed {
			consumed = start + channelInfoLength
		}
	} else {
		c.WriteChannelInfo = []byte{}
	}

	return consumed, nil
}
