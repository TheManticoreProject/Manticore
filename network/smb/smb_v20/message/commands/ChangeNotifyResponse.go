package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
)

const (
	// ChangeNotifyResponseStructureSize is the fixed StructureSize value the server
	// MUST set for an SMB2 CHANGE_NOTIFY Response (8-byte fixed part + 1).
	ChangeNotifyResponseStructureSize = 9

	// changeNotifyResponseFixedSize is the size, in bytes, of the fixed portion of
	// the body that precedes the variable output Buffer.
	changeNotifyResponseFixedSize = 8
)

// ChangeNotifyResponse is the SMB2 CHANGE_NOTIFY Response body, sent by the server
// with a buffer of FILE_NOTIFY_INFORMATION change records.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/14f9d050-27b2-49df-b009-54e08e8bf7b5
type ChangeNotifyResponse struct {
	command_interface.Command

	// OutputBuffer holds the FILE_NOTIFY_INFORMATION records, carried verbatim.
	// OutputBufferOffset/Length are computed on marshal.
	OutputBuffer []byte
}

// NewChangeNotifyResponse creates a new SMB2 CHANGE_NOTIFY Response.
func NewChangeNotifyResponse() *ChangeNotifyResponse {
	c := &ChangeNotifyResponse{OutputBuffer: []byte{}}
	c.SetCommandCode(codes.SMB2_CHANGE_NOTIFY)
	c.StructureSize = ChangeNotifyResponseStructureSize
	return c
}

// Marshal serializes the CHANGE_NOTIFY Response body.
func (c *ChangeNotifyResponse) Marshal() ([]byte, error) {
	buf := make([]byte, changeNotifyResponseFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], ChangeNotifyResponseStructureSize)
	if len(c.OutputBuffer) > 0 {
		binary.LittleEndian.PutUint16(buf[2:4], uint16(header.SMB2_HEADER_SIZE+changeNotifyResponseFixedSize))
	}
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(c.OutputBuffer)))

	if len(c.OutputBuffer) > 0 {
		buf = append(buf, c.OutputBuffer...)
	} else {
		// Keep a non-empty Buffer to match the StructureSize off-by-one convention.
		buf = append(buf, 0x00)
	}

	return buf, nil
}

// Unmarshal deserializes the CHANGE_NOTIFY Response body.
func (c *ChangeNotifyResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < changeNotifyResponseFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 CHANGE_NOTIFY Response: have %d bytes, need at least %d", len(data), changeNotifyResponseFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	outputOffset := int(binary.LittleEndian.Uint16(data[2:4]))
	outputLength := int(binary.LittleEndian.Uint32(data[4:8]))

	consumed := changeNotifyResponseFixedSize
	if outputLength > 0 {
		start := outputOffset - header.SMB2_HEADER_SIZE
		if start < changeNotifyResponseFixedSize || start+outputLength > len(data) {
			return 0, fmt.Errorf("SMB2 CHANGE_NOTIFY Response output out of bounds: offset %d length %d", outputOffset, outputLength)
		}
		c.OutputBuffer = make([]byte, outputLength)
		copy(c.OutputBuffer, data[start:start+outputLength])
		consumed = start + outputLength
	} else {
		c.OutputBuffer = []byte{}
	}

	return consumed, nil
}
