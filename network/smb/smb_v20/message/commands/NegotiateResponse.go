package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

const (
	// NegotiateResponseStructureSize is the fixed StructureSize value the server
	// MUST set for an SMB2 NEGOTIATE Response (64-byte fixed part + 1).
	NegotiateResponseStructureSize = 65

	// negotiateResponseFixedSize is the size, in bytes, of the fixed portion of
	// the body that precedes the variable security Buffer.
	negotiateResponseFixedSize = 64
)

// NegotiateResponse is the SMB2 NEGOTIATE Response body, sent by the server to
// notify the client of the preferred common dialect and server capabilities.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/63abf97c-0d09-47e2-88d6-6bfa552949a5
type NegotiateResponse struct {
	command_interface.Command

	// SecurityMode (2 bytes): Whether SMB signing is enabled or required at the server.
	SecurityMode securitymode.SecurityMode

	// DialectRevision (2 bytes): The preferred common dialect, or the wildcard revision.
	DialectRevision dialects.Dialect

	// NegotiateContextCount (2 bytes): Number of negotiate contexts (SMB 3.1.1 only; else reserved).
	NegotiateContextCount types.USHORT

	// ServerGuid (16 bytes): A GUID identifying the server.
	ServerGuid [16]byte

	// Capabilities (4 bytes): Server protocol capabilities.
	Capabilities capabilities.Capabilities

	// MaxTransactSize (4 bytes): Max buffer size for QUERY_INFO/QUERY_DIRECTORY/SET_INFO/CHANGE_NOTIFY.
	MaxTransactSize types.ULONG

	// MaxReadSize (4 bytes): Max READ length the server accepts.
	MaxReadSize types.ULONG

	// MaxWriteSize (4 bytes): Max WRITE length the server accepts.
	MaxWriteSize types.ULONG

	// SystemTime (8 bytes): Server system time, as a 64-bit FILETIME value.
	SystemTime types.UINT64

	// ServerStartTime (8 bytes): Server start time, as a 64-bit FILETIME value.
	ServerStartTime types.UINT64

	// NegotiateContextOffset (4 bytes): Offset to the first negotiate context (SMB 3.1.1 only; else reserved).
	NegotiateContextOffset types.ULONG

	// SecurityBuffer (variable): The GSS security token. SecurityBufferOffset and
	// SecurityBufferLength are computed from this on marshal.
	SecurityBuffer []byte
}

// NewNegotiateResponse creates a new SMB2 NEGOTIATE Response.
func NewNegotiateResponse() *NegotiateResponse {
	c := &NegotiateResponse{SecurityBuffer: []byte{}}
	c.SetCommandCode(codes.SMB2_NEGOTIATE)
	c.StructureSize = NegotiateResponseStructureSize
	return c
}

// Marshal serializes the NEGOTIATE Response body.
func (c *NegotiateResponse) Marshal() ([]byte, error) {
	buf := make([]byte, negotiateResponseFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], NegotiateResponseStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], uint16(c.SecurityMode))
	binary.LittleEndian.PutUint16(buf[4:6], uint16(c.DialectRevision))
	binary.LittleEndian.PutUint16(buf[6:8], c.NegotiateContextCount)
	copy(buf[8:24], c.ServerGuid[:])
	binary.LittleEndian.PutUint32(buf[24:28], uint32(c.Capabilities))
	binary.LittleEndian.PutUint32(buf[28:32], c.MaxTransactSize)
	binary.LittleEndian.PutUint32(buf[32:36], c.MaxReadSize)
	binary.LittleEndian.PutUint32(buf[36:40], c.MaxWriteSize)
	binary.LittleEndian.PutUint64(buf[40:48], c.SystemTime)
	binary.LittleEndian.PutUint64(buf[48:56], c.ServerStartTime)

	// SecurityBufferOffset is measured from the start of the SMB2 header.
	if len(c.SecurityBuffer) > 0 {
		binary.LittleEndian.PutUint16(buf[56:58], uint16(header.SMB2_HEADER_SIZE+negotiateResponseFixedSize))
	}
	binary.LittleEndian.PutUint16(buf[58:60], uint16(len(c.SecurityBuffer)))
	binary.LittleEndian.PutUint32(buf[60:64], c.NegotiateContextOffset)

	buf = append(buf, c.SecurityBuffer...)
	return buf, nil
}

// Unmarshal deserializes the NEGOTIATE Response body. The input begins at the
// command body (immediately after the 64-byte SMB2 header), and the
// SecurityBufferOffset it carries is header-relative.
func (c *NegotiateResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < negotiateResponseFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 NEGOTIATE Response: have %d bytes, need at least %d", len(data), negotiateResponseFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.SecurityMode = securitymode.SecurityMode(binary.LittleEndian.Uint16(data[2:4]))
	c.DialectRevision = dialects.Dialect(binary.LittleEndian.Uint16(data[4:6]))
	c.NegotiateContextCount = binary.LittleEndian.Uint16(data[6:8])
	copy(c.ServerGuid[:], data[8:24])
	c.Capabilities = capabilities.Capabilities(binary.LittleEndian.Uint32(data[24:28]))
	c.MaxTransactSize = binary.LittleEndian.Uint32(data[28:32])
	c.MaxReadSize = binary.LittleEndian.Uint32(data[32:36])
	c.MaxWriteSize = binary.LittleEndian.Uint32(data[36:40])
	c.SystemTime = binary.LittleEndian.Uint64(data[40:48])
	c.ServerStartTime = binary.LittleEndian.Uint64(data[48:56])
	securityBufferOffset := int(binary.LittleEndian.Uint16(data[56:58]))
	securityBufferLength := int(binary.LittleEndian.Uint16(data[58:60]))
	c.NegotiateContextOffset = binary.LittleEndian.Uint32(data[60:64])

	consumed := negotiateResponseFixedSize
	if securityBufferLength > 0 {
		start := securityBufferOffset - header.SMB2_HEADER_SIZE
		if start < negotiateResponseFixedSize || start+securityBufferLength > len(data) {
			return 0, fmt.Errorf("SMB2 NEGOTIATE Response security buffer out of bounds: offset %d length %d", securityBufferOffset, securityBufferLength)
		}
		c.SecurityBuffer = make([]byte, securityBufferLength)
		copy(c.SecurityBuffer, data[start:start+securityBufferLength])
		consumed = start + securityBufferLength
	} else {
		c.SecurityBuffer = []byte{}
	}

	return consumed, nil
}
