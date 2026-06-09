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
	// CreateRequestStructureSize is the fixed StructureSize value the client MUST
	// set for an SMB2 CREATE Request (56-byte fixed part + 1).
	CreateRequestStructureSize = 57

	// createRequestFixedSize is the size, in bytes, of the fixed portion of the
	// body that precedes the variable Buffer (name + create contexts).
	createRequestFixedSize = 56
)

// SMB2 oplock levels (CREATE RequestedOplockLevel / Response OplockLevel).
const (
	SMB2_OPLOCK_LEVEL_NONE      = 0x00
	SMB2_OPLOCK_LEVEL_II        = 0x01
	SMB2_OPLOCK_LEVEL_EXCLUSIVE = 0x08
	SMB2_OPLOCK_LEVEL_BATCH     = 0x09
	SMB2_OPLOCK_LEVEL_LEASE     = 0xFF
)

// CreateRequest is the SMB2 CREATE Request body, sent by a client to create or
// open a file, named pipe, or printer.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/e8fb45c1-a03d-44ca-b7ae-47385cfd7997
type CreateRequest struct {
	command_interface.Command

	// SecurityFlags (1 byte): Reserved; the client MUST set this to 0.
	SecurityFlags types.UCHAR

	// RequestedOplockLevel (1 byte): The requested oplock level.
	RequestedOplockLevel types.UCHAR

	// ImpersonationLevel (4 bytes): The requested impersonation level.
	ImpersonationLevel types.ULONG

	// SmbCreateFlags (8 bytes): Reserved; the client SHOULD set this to 0.
	SmbCreateFlags types.UINT64

	// Reserved (8 bytes): Reserved; ignored by the server.
	Reserved types.UINT64

	// DesiredAccess (4 bytes): The level of access required.
	DesiredAccess types.ULONG

	// FileAttributes (4 bytes): The file attributes (MS-FSCC 2.6).
	FileAttributes types.ULONG

	// ShareAccess (4 bytes): The sharing mode for the open.
	ShareAccess types.ULONG

	// CreateDisposition (4 bytes): The action to take if the file exists.
	CreateDisposition types.ULONG

	// CreateOptions (4 bytes): The options to apply when creating/opening.
	CreateOptions types.ULONG

	// Name is the file name relative to the share (Unicode on the wire). An empty
	// name opens the root of the share. NameOffset/NameLength are computed on marshal.
	Name string

	// CreateContexts is the raw, 8-byte-aligned list of SMB2_CREATE_CONTEXT
	// structures, carried verbatim. CreateContextsOffset/Length are computed on
	// marshal.
	CreateContexts []byte
}

// NewCreateRequest creates a new SMB2 CREATE Request.
func NewCreateRequest() *CreateRequest {
	c := &CreateRequest{CreateContexts: []byte{}}
	c.SetCommandCode(codes.SMB2_CREATE)
	c.StructureSize = CreateRequestStructureSize
	return c
}

// Marshal serializes the CREATE Request body.
func (c *CreateRequest) Marshal() ([]byte, error) {
	nameBytes := utf16LEEncode(c.Name)

	buf := make([]byte, createRequestFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], CreateRequestStructureSize)
	buf[2] = byte(c.SecurityFlags)
	buf[3] = byte(c.RequestedOplockLevel)
	binary.LittleEndian.PutUint32(buf[4:8], c.ImpersonationLevel)
	binary.LittleEndian.PutUint64(buf[8:16], c.SmbCreateFlags)
	binary.LittleEndian.PutUint64(buf[16:24], c.Reserved)
	binary.LittleEndian.PutUint32(buf[24:28], c.DesiredAccess)
	binary.LittleEndian.PutUint32(buf[28:32], c.FileAttributes)
	binary.LittleEndian.PutUint32(buf[32:36], c.ShareAccess)
	binary.LittleEndian.PutUint32(buf[36:40], c.CreateDisposition)
	binary.LittleEndian.PutUint32(buf[40:44], c.CreateOptions)

	// The name (if any) begins immediately after the fixed part; offsets are
	// measured from the start of the SMB2 header.
	variable := []byte{}
	if len(nameBytes) > 0 {
		binary.LittleEndian.PutUint16(buf[44:46], uint16(header.SMB2_HEADER_SIZE+createRequestFixedSize))
		binary.LittleEndian.PutUint16(buf[46:48], uint16(len(nameBytes)))
		variable = append(variable, nameBytes...)
	}

	if len(c.CreateContexts) > 0 {
		// Create contexts must start on an 8-byte boundary relative to the header.
		posInHeader := header.SMB2_HEADER_SIZE + createRequestFixedSize + len(variable)
		pad := align8(posInHeader) - posInHeader
		variable = append(variable, make([]byte, pad)...)
		binary.LittleEndian.PutUint32(buf[48:52], uint32(header.SMB2_HEADER_SIZE+createRequestFixedSize+len(variable)))
		binary.LittleEndian.PutUint32(buf[52:56], uint32(len(c.CreateContexts)))
		variable = append(variable, c.CreateContexts...)
	}

	// The Buffer field MUST be at least one byte in length.
	if len(variable) == 0 {
		variable = []byte{0x00}
	}

	return append(buf, variable...), nil
}

// Unmarshal deserializes the CREATE Request body.
func (c *CreateRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < createRequestFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 CREATE Request: have %d bytes, need at least %d", len(data), createRequestFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.SecurityFlags = data[2]
	c.RequestedOplockLevel = data[3]
	c.ImpersonationLevel = binary.LittleEndian.Uint32(data[4:8])
	c.SmbCreateFlags = binary.LittleEndian.Uint64(data[8:16])
	c.Reserved = binary.LittleEndian.Uint64(data[16:24])
	c.DesiredAccess = binary.LittleEndian.Uint32(data[24:28])
	c.FileAttributes = binary.LittleEndian.Uint32(data[28:32])
	c.ShareAccess = binary.LittleEndian.Uint32(data[32:36])
	c.CreateDisposition = binary.LittleEndian.Uint32(data[36:40])
	c.CreateOptions = binary.LittleEndian.Uint32(data[40:44])
	nameOffset := int(binary.LittleEndian.Uint16(data[44:46]))
	nameLength := int(binary.LittleEndian.Uint16(data[46:48]))
	ccOffset := int(binary.LittleEndian.Uint32(data[48:52]))
	ccLength := int(binary.LittleEndian.Uint32(data[52:56]))

	consumed := createRequestFixedSize
	if nameLength > 0 {
		start := nameOffset - header.SMB2_HEADER_SIZE
		if start < createRequestFixedSize || start+nameLength > len(data) {
			return 0, fmt.Errorf("SMB2 CREATE Request name out of bounds: offset %d length %d", nameOffset, nameLength)
		}
		c.Name = utf16LEDecode(data[start : start+nameLength])
		if start+nameLength > consumed {
			consumed = start + nameLength
		}
	} else {
		c.Name = ""
	}

	if ccLength > 0 {
		start := ccOffset - header.SMB2_HEADER_SIZE
		if start < createRequestFixedSize || start+ccLength > len(data) {
			return 0, fmt.Errorf("SMB2 CREATE Request create-contexts out of bounds: offset %d length %d", ccOffset, ccLength)
		}
		c.CreateContexts = make([]byte, ccLength)
		copy(c.CreateContexts, data[start:start+ccLength])
		if start+ccLength > consumed {
			consumed = start + ccLength
		}
	} else {
		c.CreateContexts = []byte{}
	}

	return consumed, nil
}

// align8 rounds n up to the next multiple of 8.
func align8(n int) int {
	return (n + 7) &^ 7
}
