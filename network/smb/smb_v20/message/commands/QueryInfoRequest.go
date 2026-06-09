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
	// QueryInfoRequestStructureSize is the fixed StructureSize value the client
	// MUST set for an SMB2 QUERY_INFO Request (40-byte fixed part + 1).
	QueryInfoRequestStructureSize = 41

	// queryInfoRequestFixedSize is the size, in bytes, of the fixed portion of the
	// body that precedes the variable input Buffer.
	queryInfoRequestFixedSize = 40
)

// SMB2 QUERY_INFO / SET_INFO information types (InfoType field).
const (
	SMB2_0_INFO_FILE       = 0x01
	SMB2_0_INFO_FILESYSTEM = 0x02
	SMB2_0_INFO_SECURITY   = 0x03
	SMB2_0_INFO_QUOTA      = 0x04
)

// QueryInfoRequest is the SMB2 QUERY_INFO Request body, sent by a client to query
// information about a file, named pipe, or volume identified by FileId.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/d623b2f7-a5cd-4639-8cc9-71fa7d9f9ba9
type QueryInfoRequest struct {
	command_interface.Command

	// InfoType (1 byte): The class of information requested (file/filesystem/security/quota).
	InfoType types.UCHAR

	// FileInfoClass (1 byte): The specific information class within InfoType.
	FileInfoClass types.UCHAR

	// OutputBufferLength (4 bytes): The maximum number of response bytes desired.
	OutputBufferLength types.ULONG

	// Reserved (2 bytes): The client MUST set this to 0.
	Reserved types.USHORT

	// AdditionalInformation (4 bytes): Extra query parameters (security info bits, EA index).
	AdditionalInformation types.ULONG

	// Flags (4 bytes): EA-scan flags for FileFullEaInformation; 0 otherwise.
	Flags types.ULONG

	// FileId (16 bytes): The file/pipe/volume to query.
	FileId types.SMB2_FILEID

	// Input is the optional input buffer (quota info / EA list).
	// InputBufferOffset/InputBufferLength are computed on marshal.
	Input []byte
}

// NewQueryInfoRequest creates a new SMB2 QUERY_INFO Request.
func NewQueryInfoRequest() *QueryInfoRequest {
	c := &QueryInfoRequest{Input: []byte{}}
	c.SetCommandCode(codes.SMB2_QUERY_INFO)
	c.StructureSize = QueryInfoRequestStructureSize
	return c
}

// Marshal serializes the QUERY_INFO Request body.
func (c *QueryInfoRequest) Marshal() ([]byte, error) {
	buf := make([]byte, queryInfoRequestFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], QueryInfoRequestStructureSize)
	buf[2] = byte(c.InfoType)
	buf[3] = byte(c.FileInfoClass)
	binary.LittleEndian.PutUint32(buf[4:8], c.OutputBufferLength)
	if len(c.Input) > 0 {
		binary.LittleEndian.PutUint16(buf[8:10], uint16(header.SMB2_HEADER_SIZE+queryInfoRequestFixedSize))
	}
	binary.LittleEndian.PutUint16(buf[10:12], c.Reserved)
	binary.LittleEndian.PutUint32(buf[12:16], uint32(len(c.Input)))
	binary.LittleEndian.PutUint32(buf[16:20], c.AdditionalInformation)
	binary.LittleEndian.PutUint32(buf[20:24], c.Flags)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	copy(buf[24:40], fileId)

	if len(c.Input) > 0 {
		buf = append(buf, c.Input...)
	} else {
		// Keep a non-empty Buffer to match the StructureSize off-by-one convention.
		buf = append(buf, 0x00)
	}

	return buf, nil
}

// Unmarshal deserializes the QUERY_INFO Request body.
func (c *QueryInfoRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < queryInfoRequestFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 QUERY_INFO Request: have %d bytes, need at least %d", len(data), queryInfoRequestFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.InfoType = data[2]
	c.FileInfoClass = data[3]
	c.OutputBufferLength = binary.LittleEndian.Uint32(data[4:8])
	inputOffset := int(binary.LittleEndian.Uint16(data[8:10]))
	c.Reserved = binary.LittleEndian.Uint16(data[10:12])
	inputLength := int(binary.LittleEndian.Uint32(data[12:16]))
	c.AdditionalInformation = binary.LittleEndian.Uint32(data[16:20])
	c.Flags = binary.LittleEndian.Uint32(data[20:24])
	if _, err := c.FileId.Unmarshal(data[24:40]); err != nil {
		return 0, err
	}

	consumed := queryInfoRequestFixedSize
	if inputLength > 0 {
		start := inputOffset - header.SMB2_HEADER_SIZE
		if start < queryInfoRequestFixedSize || start+inputLength > len(data) {
			return 0, fmt.Errorf("SMB2 QUERY_INFO Request input out of bounds: offset %d length %d", inputOffset, inputLength)
		}
		c.Input = make([]byte, inputLength)
		copy(c.Input, data[start:start+inputLength])
		consumed = start + inputLength
	} else {
		c.Input = []byte{}
	}

	return consumed, nil
}
