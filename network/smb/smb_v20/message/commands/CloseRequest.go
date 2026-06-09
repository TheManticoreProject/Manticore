package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

const (
	// CloseRequestStructureSize is the fixed StructureSize value for an SMB2 CLOSE Request.
	CloseRequestStructureSize = 24

	// SMB2_CLOSE_FLAG_POSTQUERY_ATTRIB requests that the server return file
	// attributes in the CLOSE Response.
	SMB2_CLOSE_FLAG_POSTQUERY_ATTRIB = 0x0001
)

// CloseRequest is the SMB2 CLOSE Request body, sent by the client to close an
// open identified by FileId.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/f84053b0-bcb2-4f85-9717-536dae2b02bd
type CloseRequest struct {
	command_interface.Command

	// Flags (2 bytes): If SMB2_CLOSE_FLAG_POSTQUERY_ATTRIB is set, the server
	// returns file attributes in the response.
	Flags types.USHORT

	// Reserved (4 bytes): The client MUST set this to 0.
	Reserved types.ULONG

	// FileId (16 bytes): The open to be closed.
	FileId types.SMB2_FILEID
}

// NewCloseRequest creates a new SMB2 CLOSE Request.
func NewCloseRequest() *CloseRequest {
	c := &CloseRequest{}
	c.SetCommandCode(codes.SMB2_CLOSE)
	c.StructureSize = CloseRequestStructureSize
	return c
}

// Marshal serializes the CLOSE Request body.
func (c *CloseRequest) Marshal() ([]byte, error) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint16(buf[0:2], CloseRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Flags)
	binary.LittleEndian.PutUint32(buf[4:8], c.Reserved)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	buf = append(buf, fileId...)
	return buf, nil
}

// Unmarshal deserializes the CLOSE Request body.
func (c *CloseRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < 24 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 CLOSE Request: have %d bytes, need 24", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Flags = binary.LittleEndian.Uint16(data[2:4])
	c.Reserved = binary.LittleEndian.Uint32(data[4:8])
	if _, err := c.FileId.Unmarshal(data[8:24]); err != nil {
		return 0, err
	}
	return 24, nil
}
