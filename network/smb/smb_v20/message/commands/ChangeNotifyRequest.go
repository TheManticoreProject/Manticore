package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

const (
	// ChangeNotifyRequestStructureSize is the fixed StructureSize value for an SMB2
	// CHANGE_NOTIFY Request. The request has no variable buffer, so this equals the
	// body size.
	ChangeNotifyRequestStructureSize = 32

	// SMB2_WATCH_TREE requests that the directory be monitored recursively.
	SMB2_WATCH_TREE = 0x0001
)

// ChangeNotifyRequest is the SMB2 CHANGE_NOTIFY Request body, sent by the client
// to be notified of changes to the directory identified by FileId.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/598f395a-e7a2-4cc8-afb3-ccb30dd2df7c
type ChangeNotifyRequest struct {
	command_interface.Command

	// Flags (2 bytes): SMB2_WATCH_TREE to monitor the directory recursively.
	Flags types.USHORT

	// OutputBufferLength (4 bytes): The maximum number of response bytes desired.
	OutputBufferLength types.ULONG

	// FileId (16 bytes): The directory to monitor.
	FileId types.SMB2_FILEID

	// CompletionFilter (4 bytes): The types of changes to monitor (FILE_NOTIFY_CHANGE_*).
	CompletionFilter types.ULONG

	// Reserved (4 bytes): The client MUST set this to 0.
	Reserved types.ULONG
}

// NewChangeNotifyRequest creates a new SMB2 CHANGE_NOTIFY Request.
func NewChangeNotifyRequest() *ChangeNotifyRequest {
	c := &ChangeNotifyRequest{}
	c.SetCommandCode(codes.SMB2_CHANGE_NOTIFY)
	c.StructureSize = ChangeNotifyRequestStructureSize
	return c
}

// Marshal serializes the CHANGE_NOTIFY Request body.
func (c *ChangeNotifyRequest) Marshal() ([]byte, error) {
	buf := make([]byte, ChangeNotifyRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[0:2], ChangeNotifyRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Flags)
	binary.LittleEndian.PutUint32(buf[4:8], c.OutputBufferLength)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	copy(buf[8:24], fileId)

	binary.LittleEndian.PutUint32(buf[24:28], c.CompletionFilter)
	binary.LittleEndian.PutUint32(buf[28:32], c.Reserved)
	return buf, nil
}

// Unmarshal deserializes the CHANGE_NOTIFY Request body.
func (c *ChangeNotifyRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < ChangeNotifyRequestStructureSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 CHANGE_NOTIFY Request: have %d bytes, need %d", len(data), ChangeNotifyRequestStructureSize)
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Flags = binary.LittleEndian.Uint16(data[2:4])
	c.OutputBufferLength = binary.LittleEndian.Uint32(data[4:8])
	if _, err := c.FileId.Unmarshal(data[8:24]); err != nil {
		return 0, err
	}
	c.CompletionFilter = binary.LittleEndian.Uint32(data[24:28])
	c.Reserved = binary.LittleEndian.Uint32(data[28:32])
	return ChangeNotifyRequestStructureSize, nil
}
