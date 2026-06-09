package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
)

const (
	// QueryDirectoryResponseStructureSize is the fixed StructureSize value the
	// server MUST set for an SMB2 QUERY_DIRECTORY Response (8-byte fixed part + 1).
	QueryDirectoryResponseStructureSize = 9

	// queryDirectoryResponseFixedSize is the size, in bytes, of the fixed portion
	// of the body that precedes the variable output Buffer.
	queryDirectoryResponseFixedSize = 8
)

// QueryDirectoryResponse is the SMB2 QUERY_DIRECTORY Response body, sent by the
// server with a buffer of directory entries.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/4f75351b-048c-4a0c-9ea3-addd55a71956
type QueryDirectoryResponse struct {
	command_interface.Command

	// OutputBuffer holds the directory-entry information (MS-FSCC structures),
	// carried verbatim. OutputBufferOffset/Length are computed on marshal.
	OutputBuffer []byte
}

// NewQueryDirectoryResponse creates a new SMB2 QUERY_DIRECTORY Response.
func NewQueryDirectoryResponse() *QueryDirectoryResponse {
	c := &QueryDirectoryResponse{OutputBuffer: []byte{}}
	c.SetCommandCode(codes.SMB2_QUERY_DIRECTORY)
	c.StructureSize = QueryDirectoryResponseStructureSize
	return c
}

// Marshal serializes the QUERY_DIRECTORY Response body.
func (c *QueryDirectoryResponse) Marshal() ([]byte, error) {
	buf := make([]byte, queryDirectoryResponseFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], QueryDirectoryResponseStructureSize)
	if len(c.OutputBuffer) > 0 {
		binary.LittleEndian.PutUint16(buf[2:4], uint16(header.SMB2_HEADER_SIZE+queryDirectoryResponseFixedSize))
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

// Unmarshal deserializes the QUERY_DIRECTORY Response body.
func (c *QueryDirectoryResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < queryDirectoryResponseFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 QUERY_DIRECTORY Response: have %d bytes, need at least %d", len(data), queryDirectoryResponseFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	outputOffset := int(binary.LittleEndian.Uint16(data[2:4]))
	outputLength := int(binary.LittleEndian.Uint32(data[4:8]))

	consumed := queryDirectoryResponseFixedSize
	if outputLength > 0 {
		start := outputOffset - header.SMB2_HEADER_SIZE
		if start < queryDirectoryResponseFixedSize || start+outputLength > len(data) {
			return 0, fmt.Errorf("SMB2 QUERY_DIRECTORY Response output out of bounds: offset %d length %d", outputOffset, outputLength)
		}
		c.OutputBuffer = make([]byte, outputLength)
		copy(c.OutputBuffer, data[start:start+outputLength])
		consumed = start + outputLength
	} else {
		c.OutputBuffer = []byte{}
	}

	return consumed, nil
}
