package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
)

// ErrorResponseStructureSize is the fixed StructureSize value for an SMB2 ERROR Response.
const ErrorResponseStructureSize = 9

// ErrorResponse is the SMB2 ERROR Response body, sent in place of a command's
// own response when the request failed. It carries the failed command's code.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/d4da8b67-c180-47e3-ba7a-d24214ac4aaa
type ErrorResponse struct {
	command_interface.Command

	// ErrorContextCount (1 byte): The number of error contexts (SMB 3.1.1 only).
	ErrorContextCount uint8

	// ErrorData (variable): Command-specific error data. An empty one is sent as
	// a single zero byte.
	ErrorData []byte
}

// NewErrorResponse creates an SMB2 ERROR Response for the given command.
func NewErrorResponse(command codes.CommandCode) *ErrorResponse {
	c := &ErrorResponse{}
	c.SetCommandCode(command)
	c.StructureSize = ErrorResponseStructureSize
	return c
}

// Marshal serializes the ERROR Response body.
func (c *ErrorResponse) Marshal() ([]byte, error) {
	buf := make([]byte, 8, 8+max(len(c.ErrorData), 1))
	binary.LittleEndian.PutUint16(buf[0:2], ErrorResponseStructureSize)
	buf[2] = c.ErrorContextCount
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(c.ErrorData)))
	if len(c.ErrorData) == 0 {
		return append(buf, 0x00), nil
	}
	return append(buf, c.ErrorData...), nil
}

// Unmarshal deserializes the ERROR Response body.
func (c *ErrorResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 ERROR Response: have %d bytes, need at least 8", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.ErrorContextCount = data[2]
	byteCount := int(binary.LittleEndian.Uint32(data[4:8]))
	if 8+byteCount > len(data) {
		return 0, fmt.Errorf("SMB2 ERROR Response data out of bounds: %d bytes declared, %d available", byteCount, len(data)-8)
	}
	c.ErrorData = append([]byte{}, data[8:8+byteCount]...)
	return 8 + byteCount, nil
}
