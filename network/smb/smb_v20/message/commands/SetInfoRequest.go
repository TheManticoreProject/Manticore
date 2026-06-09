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
	// SetInfoRequestStructureSize is the fixed StructureSize value the client MUST
	// set for an SMB2 SET_INFO Request (32-byte fixed part + 1).
	SetInfoRequestStructureSize = 33

	// setInfoRequestFixedSize is the size, in bytes, of the fixed portion of the
	// body that precedes the variable Buffer.
	setInfoRequestFixedSize = 32
)

// SetInfoRequest is the SMB2 SET_INFO Request body, sent by a client to set
// information on the file, named pipe, or volume identified by FileId. The
// InfoType values are shared with QUERY_INFO (SMB2_0_INFO_*).
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/ee9614c4-be54-4a3c-98f1-769a7032a0e4
type SetInfoRequest struct {
	command_interface.Command

	// InfoType (1 byte): The class of information being set (file/filesystem/security/quota).
	InfoType types.UCHAR

	// FileInfoClass (1 byte): The specific information class within InfoType.
	FileInfoClass types.UCHAR

	// AdditionalInformation (4 bytes): Security-info bits when setting security; 0 otherwise.
	AdditionalInformation types.ULONG

	// Reserved (2 bytes): The client MUST set this to 0.
	Reserved types.USHORT

	// FileId (16 bytes): The file/pipe/volume on which to set the information.
	FileId types.SMB2_FILEID

	// Buffer is the information to set (MS-FSCC / security-descriptor structures).
	// BufferOffset/BufferLength are computed on marshal.
	Buffer []byte
}

// NewSetInfoRequest creates a new SMB2 SET_INFO Request.
func NewSetInfoRequest() *SetInfoRequest {
	c := &SetInfoRequest{Buffer: []byte{}}
	c.SetCommandCode(codes.SMB2_SET_INFO)
	c.StructureSize = SetInfoRequestStructureSize
	return c
}

// Marshal serializes the SET_INFO Request body.
func (c *SetInfoRequest) Marshal() ([]byte, error) {
	buf := make([]byte, setInfoRequestFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], SetInfoRequestStructureSize)
	buf[2] = byte(c.InfoType)
	buf[3] = byte(c.FileInfoClass)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(c.Buffer)))
	if len(c.Buffer) > 0 {
		binary.LittleEndian.PutUint16(buf[8:10], uint16(header.SMB2_HEADER_SIZE+setInfoRequestFixedSize))
	}
	binary.LittleEndian.PutUint16(buf[10:12], c.Reserved)
	binary.LittleEndian.PutUint32(buf[12:16], c.AdditionalInformation)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	copy(buf[16:32], fileId)

	if len(c.Buffer) > 0 {
		buf = append(buf, c.Buffer...)
	} else {
		// Keep a non-empty Buffer to match the StructureSize off-by-one convention.
		buf = append(buf, 0x00)
	}

	return buf, nil
}

// Unmarshal deserializes the SET_INFO Request body.
func (c *SetInfoRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < setInfoRequestFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 SET_INFO Request: have %d bytes, need at least %d", len(data), setInfoRequestFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.InfoType = data[2]
	c.FileInfoClass = data[3]
	bufferLength := int(binary.LittleEndian.Uint32(data[4:8]))
	bufferOffset := int(binary.LittleEndian.Uint16(data[8:10]))
	c.Reserved = binary.LittleEndian.Uint16(data[10:12])
	c.AdditionalInformation = binary.LittleEndian.Uint32(data[12:16])
	if _, err := c.FileId.Unmarshal(data[16:32]); err != nil {
		return 0, err
	}

	consumed := setInfoRequestFixedSize
	if bufferLength > 0 {
		start := bufferOffset - header.SMB2_HEADER_SIZE
		if start < setInfoRequestFixedSize || start+bufferLength > len(data) {
			return 0, fmt.Errorf("SMB2 SET_INFO Request buffer out of bounds: offset %d length %d", bufferOffset, bufferLength)
		}
		c.Buffer = make([]byte, bufferLength)
		copy(c.Buffer, data[start:start+bufferLength])
		consumed = start + bufferLength
	} else {
		c.Buffer = []byte{}
	}

	return consumed, nil
}
