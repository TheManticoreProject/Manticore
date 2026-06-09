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
	// CreateResponseStructureSize is the fixed StructureSize value the server MUST
	// set for an SMB2 CREATE Response (88-byte fixed part + 1).
	CreateResponseStructureSize = 89

	// createResponseFixedSize is the size, in bytes, of the fixed portion of the
	// body that precedes the variable create-contexts Buffer.
	createResponseFixedSize = 88
)

// CreateResponse is the SMB2 CREATE Response body, sent by the server to report
// the result of an SMB2 CREATE Request and return the open's FileId.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/d166aa9e-0b53-410e-b35e-3933d8131927
type CreateResponse struct {
	command_interface.Command

	// OplockLevel (1 byte): The oplock level granted for this open.
	OplockLevel types.UCHAR

	// Flags (1 byte): SMB 3.x reparse-point flag; reserved otherwise.
	Flags types.UCHAR

	// CreateAction (4 bytes): The action taken (superseded/opened/created/overwritten).
	CreateAction types.ULONG

	// CreationTime (8 bytes): When the file was created, as a 64-bit FILETIME.
	CreationTime types.UINT64

	// LastAccessTime (8 bytes): When the file was last accessed, as a 64-bit FILETIME.
	LastAccessTime types.UINT64

	// LastWriteTime (8 bytes): When data was last written, as a 64-bit FILETIME.
	LastWriteTime types.UINT64

	// ChangeTime (8 bytes): When the file was last modified, as a 64-bit FILETIME.
	ChangeTime types.UINT64

	// AllocationSize (8 bytes): The allocation size of the file, in bytes.
	AllocationSize types.UINT64

	// EndOfFile (8 bytes): The size of the file, in bytes.
	EndOfFile types.UINT64

	// FileAttributes (4 bytes): The attributes of the file.
	FileAttributes types.ULONG

	// Reserved2 (4 bytes): Reserved; the server SHOULD set this to 0.
	Reserved2 types.ULONG

	// FileId (16 bytes): The identifier of the established open.
	FileId types.SMB2_FILEID

	// CreateContexts is the raw list of SMB2_CREATE_CONTEXT response structures,
	// carried verbatim. CreateContextsOffset/Length are computed on marshal.
	CreateContexts []byte
}

// NewCreateResponse creates a new SMB2 CREATE Response.
func NewCreateResponse() *CreateResponse {
	c := &CreateResponse{CreateContexts: []byte{}}
	c.SetCommandCode(codes.SMB2_CREATE)
	c.StructureSize = CreateResponseStructureSize
	return c
}

// Marshal serializes the CREATE Response body.
func (c *CreateResponse) Marshal() ([]byte, error) {
	buf := make([]byte, createResponseFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], CreateResponseStructureSize)
	buf[2] = byte(c.OplockLevel)
	buf[3] = byte(c.Flags)
	binary.LittleEndian.PutUint32(buf[4:8], c.CreateAction)
	binary.LittleEndian.PutUint64(buf[8:16], c.CreationTime)
	binary.LittleEndian.PutUint64(buf[16:24], c.LastAccessTime)
	binary.LittleEndian.PutUint64(buf[24:32], c.LastWriteTime)
	binary.LittleEndian.PutUint64(buf[32:40], c.ChangeTime)
	binary.LittleEndian.PutUint64(buf[40:48], c.AllocationSize)
	binary.LittleEndian.PutUint64(buf[48:56], c.EndOfFile)
	binary.LittleEndian.PutUint32(buf[56:60], c.FileAttributes)
	binary.LittleEndian.PutUint32(buf[60:64], c.Reserved2)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	copy(buf[64:80], fileId)

	if len(c.CreateContexts) > 0 {
		binary.LittleEndian.PutUint32(buf[80:84], uint32(header.SMB2_HEADER_SIZE+createResponseFixedSize))
		binary.LittleEndian.PutUint32(buf[84:88], uint32(len(c.CreateContexts)))
		buf = append(buf, c.CreateContexts...)
	}

	return buf, nil
}

// Unmarshal deserializes the CREATE Response body.
func (c *CreateResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < createResponseFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 CREATE Response: have %d bytes, need at least %d", len(data), createResponseFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.OplockLevel = data[2]
	c.Flags = data[3]
	c.CreateAction = binary.LittleEndian.Uint32(data[4:8])
	c.CreationTime = binary.LittleEndian.Uint64(data[8:16])
	c.LastAccessTime = binary.LittleEndian.Uint64(data[16:24])
	c.LastWriteTime = binary.LittleEndian.Uint64(data[24:32])
	c.ChangeTime = binary.LittleEndian.Uint64(data[32:40])
	c.AllocationSize = binary.LittleEndian.Uint64(data[40:48])
	c.EndOfFile = binary.LittleEndian.Uint64(data[48:56])
	c.FileAttributes = binary.LittleEndian.Uint32(data[56:60])
	c.Reserved2 = binary.LittleEndian.Uint32(data[60:64])
	if _, err := c.FileId.Unmarshal(data[64:80]); err != nil {
		return 0, err
	}
	ccOffset := int(binary.LittleEndian.Uint32(data[80:84]))
	ccLength := int(binary.LittleEndian.Uint32(data[84:88]))

	consumed := createResponseFixedSize
	if ccLength > 0 {
		start := ccOffset - header.SMB2_HEADER_SIZE
		if start < createResponseFixedSize || start+ccLength > len(data) {
			return 0, fmt.Errorf("SMB2 CREATE Response create-contexts out of bounds: offset %d length %d", ccOffset, ccLength)
		}
		c.CreateContexts = make([]byte, ccLength)
		copy(c.CreateContexts, data[start:start+ccLength])
		consumed = start + ccLength
	} else {
		c.CreateContexts = []byte{}
	}

	return consumed, nil
}
