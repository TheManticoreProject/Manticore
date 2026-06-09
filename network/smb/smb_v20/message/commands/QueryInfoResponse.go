package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
)

const (
	// QueryInfoResponseStructureSize is the fixed StructureSize value the server
	// MUST set for an SMB2 QUERY_INFO Response (8-byte fixed part + 1).
	QueryInfoResponseStructureSize = 9

	// queryInfoResponseFixedSize is the size, in bytes, of the fixed portion of
	// the body that precedes the variable output Buffer.
	queryInfoResponseFixedSize = 8
)

// QueryInfoResponse is the SMB2 QUERY_INFO Response body, sent by the server with
// the requested information in its output buffer.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/3b1b3598-a898-44ca-bfac-2dcae065247f
type QueryInfoResponse struct {
	command_interface.Command

	// OutputBuffer holds the queried information (MS-FSCC / security descriptor /
	// quota structures), carried verbatim. OutputBufferOffset/Length are computed
	// on marshal.
	OutputBuffer []byte
}

// NewQueryInfoResponse creates a new SMB2 QUERY_INFO Response.
func NewQueryInfoResponse() *QueryInfoResponse {
	c := &QueryInfoResponse{OutputBuffer: []byte{}}
	c.SetCommandCode(codes.SMB2_QUERY_INFO)
	c.StructureSize = QueryInfoResponseStructureSize
	return c
}

// Marshal serializes the QUERY_INFO Response body.
func (c *QueryInfoResponse) Marshal() ([]byte, error) {
	buf := make([]byte, queryInfoResponseFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], QueryInfoResponseStructureSize)
	if len(c.OutputBuffer) > 0 {
		binary.LittleEndian.PutUint16(buf[2:4], uint16(header.SMB2_HEADER_SIZE+queryInfoResponseFixedSize))
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

// Unmarshal deserializes the QUERY_INFO Response body.
func (c *QueryInfoResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < queryInfoResponseFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 QUERY_INFO Response: have %d bytes, need at least %d", len(data), queryInfoResponseFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	outputOffset := int(binary.LittleEndian.Uint16(data[2:4]))
	outputLength := int(binary.LittleEndian.Uint32(data[4:8]))

	consumed := queryInfoResponseFixedSize
	if outputLength > 0 {
		start := outputOffset - header.SMB2_HEADER_SIZE
		if start < queryInfoResponseFixedSize || start+outputLength > len(data) {
			return 0, fmt.Errorf("SMB2 QUERY_INFO Response output out of bounds: offset %d length %d", outputOffset, outputLength)
		}
		c.OutputBuffer = make([]byte, outputLength)
		copy(c.OutputBuffer, data[start:start+outputLength])
		consumed = start + outputLength
	} else {
		c.OutputBuffer = []byte{}
	}

	return consumed, nil
}
