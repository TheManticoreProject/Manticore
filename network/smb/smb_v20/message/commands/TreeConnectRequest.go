package commands

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

const (
	// TreeConnectRequestStructureSize is the fixed StructureSize value the client
	// MUST set for an SMB2 TREE_CONNECT Request (8-byte fixed part + 1).
	TreeConnectRequestStructureSize = 9

	// treeConnectRequestFixedSize is the size, in bytes, of the fixed portion of
	// the body that precedes the variable path Buffer.
	treeConnectRequestFixedSize = 8
)

// TreeConnectRequest is the SMB2 TREE_CONNECT Request body, sent by the client to
// request access to a share. The share path is a Unicode string of the form
// "\\server\share".
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/832d2130-22e8-4afb-aafd-b30bb0901798
type TreeConnectRequest struct {
	command_interface.Command

	// Flags (2 bytes): Tree-connect flags (SMB 3.1.1 only; else reserved, 0).
	Flags types.USHORT

	// Path is the full share path. PathOffset and PathLength are computed from its
	// UTF-16LE encoding on marshal.
	Path string
}

// NewTreeConnectRequest creates a new SMB2 TREE_CONNECT Request.
func NewTreeConnectRequest() *TreeConnectRequest {
	c := &TreeConnectRequest{}
	c.SetCommandCode(codes.SMB2_TREE_CONNECT)
	c.StructureSize = TreeConnectRequestStructureSize
	return c
}

// Marshal serializes the TREE_CONNECT Request body.
func (c *TreeConnectRequest) Marshal() ([]byte, error) {
	pathBytes := utf16LEEncode(c.Path)

	buf := make([]byte, treeConnectRequestFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], TreeConnectRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Flags)

	// PathOffset is measured from the start of the SMB2 header.
	if len(pathBytes) > 0 {
		binary.LittleEndian.PutUint16(buf[4:6], uint16(header.SMB2_HEADER_SIZE+treeConnectRequestFixedSize))
	}
	binary.LittleEndian.PutUint16(buf[6:8], uint16(len(pathBytes)))

	buf = append(buf, pathBytes...)
	return buf, nil
}

// Unmarshal deserializes the TREE_CONNECT Request body.
func (c *TreeConnectRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < treeConnectRequestFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 TREE_CONNECT Request: have %d bytes, need at least %d", len(data), treeConnectRequestFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Flags = binary.LittleEndian.Uint16(data[2:4])
	pathOffset := int(binary.LittleEndian.Uint16(data[4:6]))
	pathLength := int(binary.LittleEndian.Uint16(data[6:8]))

	consumed := treeConnectRequestFixedSize
	if pathLength > 0 {
		start := pathOffset - header.SMB2_HEADER_SIZE
		if start < treeConnectRequestFixedSize || start+pathLength > len(data) {
			return 0, fmt.Errorf("SMB2 TREE_CONNECT Request path out of bounds: offset %d length %d", pathOffset, pathLength)
		}
		c.Path = utf16LEDecode(data[start : start+pathLength])
		consumed = start + pathLength
	} else {
		c.Path = ""
	}

	return consumed, nil
}

// utf16LEEncode encodes a Go string to a little-endian UTF-16 byte slice.
func utf16LEEncode(s string) []byte {
	units := utf16.Encode([]rune(s))
	buf := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(buf[i*2:], u)
	}
	return buf
}

// utf16LEDecode decodes a little-endian UTF-16 byte slice to a Go string. A
// trailing odd byte, if any, is ignored.
func utf16LEDecode(b []byte) string {
	units := make([]uint16, len(b)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return string(utf16.Decode(units))
}
