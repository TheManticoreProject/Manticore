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
	// QueryDirectoryRequestStructureSize is the fixed StructureSize value the
	// client MUST set for an SMB2 QUERY_DIRECTORY Request (32-byte fixed part + 1).
	QueryDirectoryRequestStructureSize = 33

	// queryDirectoryRequestFixedSize is the size, in bytes, of the fixed portion
	// of the body that precedes the variable search-pattern Buffer.
	queryDirectoryRequestFixedSize = 32
)

// SMB2 QUERY_DIRECTORY request flags.
const (
	SMB2_RESTART_SCANS       = 0x01
	SMB2_RETURN_SINGLE_ENTRY = 0x02
	SMB2_INDEX_SPECIFIED     = 0x04
	SMB2_REOPEN              = 0x10
)

// QueryDirectoryRequest is the SMB2 QUERY_DIRECTORY Request body, sent by the
// client to enumerate a directory identified by FileId.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/10906442-294c-46d3-8515-c277efe1f752
type QueryDirectoryRequest struct {
	command_interface.Command

	// FileInformationClass (1 byte): The format of the returned entries (MS-FSCC 2.4).
	FileInformationClass types.UCHAR

	// Flags (1 byte): How to process the enumeration (restart/single/index/reopen).
	Flags types.UCHAR

	// FileIndex (4 bytes): The resume index when SMB2_INDEX_SPECIFIED is set.
	FileIndex types.ULONG

	// FileId (16 bytes): The directory to enumerate.
	FileId types.SMB2_FILEID

	// OutputBufferLength (4 bytes): The maximum number of response bytes desired.
	OutputBufferLength types.ULONG

	// FileName is the search pattern (Unicode on the wire); empty means "*".
	// FileNameOffset/FileNameLength are computed on marshal.
	FileName string
}

// NewQueryDirectoryRequest creates a new SMB2 QUERY_DIRECTORY Request.
func NewQueryDirectoryRequest() *QueryDirectoryRequest {
	c := &QueryDirectoryRequest{}
	c.SetCommandCode(codes.SMB2_QUERY_DIRECTORY)
	c.StructureSize = QueryDirectoryRequestStructureSize
	return c
}

// Marshal serializes the QUERY_DIRECTORY Request body.
func (c *QueryDirectoryRequest) Marshal() ([]byte, error) {
	nameBytes := utf16LEEncode(c.FileName)

	buf := make([]byte, queryDirectoryRequestFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], QueryDirectoryRequestStructureSize)
	buf[2] = byte(c.FileInformationClass)
	buf[3] = byte(c.Flags)
	binary.LittleEndian.PutUint32(buf[4:8], c.FileIndex)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	copy(buf[8:24], fileId)

	if len(nameBytes) > 0 {
		binary.LittleEndian.PutUint16(buf[24:26], uint16(header.SMB2_HEADER_SIZE+queryDirectoryRequestFixedSize))
		binary.LittleEndian.PutUint16(buf[26:28], uint16(len(nameBytes)))
	}
	binary.LittleEndian.PutUint32(buf[28:32], c.OutputBufferLength)

	if len(nameBytes) > 0 {
		buf = append(buf, nameBytes...)
	} else {
		// Keep a non-empty Buffer to match the StructureSize off-by-one convention.
		buf = append(buf, 0x00)
	}

	return buf, nil
}

// Unmarshal deserializes the QUERY_DIRECTORY Request body.
func (c *QueryDirectoryRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < queryDirectoryRequestFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 QUERY_DIRECTORY Request: have %d bytes, need at least %d", len(data), queryDirectoryRequestFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.FileInformationClass = data[2]
	c.Flags = data[3]
	c.FileIndex = binary.LittleEndian.Uint32(data[4:8])
	if _, err := c.FileId.Unmarshal(data[8:24]); err != nil {
		return 0, err
	}
	nameOffset := int(binary.LittleEndian.Uint16(data[24:26]))
	nameLength := int(binary.LittleEndian.Uint16(data[26:28]))
	c.OutputBufferLength = binary.LittleEndian.Uint32(data[28:32])

	consumed := queryDirectoryRequestFixedSize
	if nameLength > 0 {
		start := nameOffset - header.SMB2_HEADER_SIZE
		if start < queryDirectoryRequestFixedSize || start+nameLength > len(data) {
			return 0, fmt.Errorf("SMB2 QUERY_DIRECTORY Request file name out of bounds: offset %d length %d", nameOffset, nameLength)
		}
		c.FileName = utf16LEDecode(data[start : start+nameLength])
		consumed = start + nameLength
	} else {
		c.FileName = ""
	}

	return consumed, nil
}
