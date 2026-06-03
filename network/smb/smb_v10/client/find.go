package client

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// smbFindFileBothDirectoryInfo is the TRANS2_FIND information level
// SMB_FIND_FILE_BOTH_DIRECTORY_INFO (MS-CIFS 2.2.8.1.7).
const smbFindFileBothDirectoryInfo = 0x0104

// fileAttributeDirectory is the SMB_EXT_FILE_ATTR bit marking a directory.
const fileAttributeDirectory = 0x00000010

// bothDirInfoFixedSize is the size of the fixed portion of an
// SMB_FIND_FILE_BOTH_DIRECTORY_INFO entry, before the variable FileName:
// NextEntryOffset(4) FileIndex(4) 4xFILETIME(32) EndOfFile(8) AllocationSize(8)
// ExtFileAttributes(4) FileNameLength(4) EaSize(4) ShortNameLength(1) Reserved(1)
// ShortName(24).
const bothDirInfoFixedSize = 94

// FileEntry is a high-level directory entry returned by ListDirectory and ListEntries.
type FileEntry struct {
	LongName    string
	ShortName   string
	Size        uint64
	Attributes  uint32
	CreatedAt   time.Time
	AccessedAt  time.Time
	ModifiedAt  time.Time
	ChangedAt   time.Time
	IsDirectory bool
}

// ListDirectory lists all entries (files and directories) in the given directory
// on the current tree. path is a directory path using backslash separators
// relative to the share root (e.g. "\", "\subdir", "\subdir\nested").
// An empty path defaults to the share root.
//
// Wire: TRANS2_FIND_FIRST2 / FIND_NEXT2 with "\path\*" pattern.
func (c *Client) ListDirectory(path string) ([]FileEntry, error) {
	if path == "" {
		path = "\\"
	}
	if path[len(path)-1] == '\\' {
		path += "*"
	} else {
		path += "\\*"
	}
	return c.ListEntries(path)
}

// ListEntries enumerates entries matching pattern in the current tree, using
// TRANS2_FIND_FIRST2 followed by as many TRANS2_FIND_NEXT2 calls as needed, and
// always closing the search with FindClose2.
//
// pattern uses SMB wildcards and backslash separators relative to the share root
// (e.g. "\*.txt", "\docs\report-*.pdf"). An empty pattern defaults to "\*".
func (c *Client) ListEntries(pattern string) ([]FileEntry, error) {
	if c.Session == nil {
		return nil, fmt.Errorf("no session established")
	}

	if pattern == "" {
		pattern = "\\*"
	}
	if pattern[0] != '\\' {
		pattern = "\\" + pattern
	}

	// TRANS2_FIND_FIRST2.
	respParams, respData, err := c.trans2(uint16(subcommands.TRANS2_FIND_FIRST2), buildFindFirst2Params(pattern), nil)
	if err != nil {
		return nil, err
	}
	if len(respParams) < 6 {
		return nil, fmt.Errorf("FIND_FIRST2 response parameters too short (%d bytes)", len(respParams))
	}

	sid := binary.LittleEndian.Uint16(respParams[0:2])
	searchCount := binary.LittleEndian.Uint16(respParams[2:4])
	endOfSearch := binary.LittleEndian.Uint16(respParams[4:6])

	entries := parseBothDirInfo(respData)

	// TRANS2_FIND_NEXT2 until the server reports end-of-search or returns nothing.
	for endOfSearch == 0 && searchCount > 0 {
		respParams, respData, err = c.trans2(uint16(subcommands.TRANS2_FIND_NEXT2), buildFindNext2Params(sid), nil)
		if err != nil {
			break
		}
		if len(respParams) < 4 {
			break
		}
		searchCount = binary.LittleEndian.Uint16(respParams[0:2])
		endOfSearch = binary.LittleEndian.Uint16(respParams[2:4])

		next := parseBothDirInfo(respData)
		if len(next) == 0 {
			break
		}
		entries = append(entries, next...)
	}

	// Always release the search handle, even on a partial enumeration.
	_ = c.findClose2(sid)

	return entries, nil
}

// trans2 issues a single SMB_COM_TRANSACTION2 carrying the given subcommand, with
// the supplied transaction parameter and data buffers, and returns the reassembled
// response parameter and data buffers. It does not implement Transaction2Secondary
// fragmentation, so the request must fit in a single message.
func (c *Client) trans2(subcommand uint16, trans2Params, trans2Data []byte) ([]byte, []byte, error) {
	msg := c.newFileIOMessage(codes.SMB_COM_TRANSACTION2)

	cmd := commands.NewTransaction2Request()
	cmd.Setup = []types.USHORT{types.USHORT(subcommand)}
	cmd.MaxParameterCount = types.USHORT(1024)
	cmd.MaxSetupCount = types.UCHAR(0)
	cmd.Flags = types.USHORT(0)
	cmd.Timeout = types.ULONG(0)

	// Bound the data the server may return by the negotiated buffer.
	maxData := 0xFFFF
	if c.Connection.Server != nil && c.Connection.Server.MaxBufferSize > 0 {
		if budget := int(c.Connection.Server.MaxBufferSize) - 512; budget > 0 && budget < maxData {
			maxData = budget
		}
	}
	cmd.MaxDataCount = types.USHORT(maxData)

	// Compute offsets. The request always carries exactly one setup word, so
	// WordCount = 14 fixed words + 1 setup word = 15.
	const wordCount = 14 + 1
	// Bytes from the start of the SMB header to the end of the Name byte: the data
	// block on the wire is ByteCount(2) + Name(1) + Pad1 + Trans2_Parameters + ...
	preParams := header.SMB_HEADER_SIZE + 1 /*WordCount*/ + 2*wordCount + 2 /*ByteCount*/ + 1 /*Name*/
	pad1 := (4 - (preParams % 4)) % 4
	parameterOffset := preParams + pad1

	cmd.TotalParameterCount = types.USHORT(len(trans2Params))
	cmd.ParameterCount = types.USHORT(len(trans2Params))
	cmd.ParameterOffset = types.USHORT(parameterOffset)
	cmd.Pad1 = make([]types.UCHAR, pad1)
	cmd.Trans2_Parameters = []types.UCHAR(trans2Params)

	if len(trans2Data) > 0 {
		afterParams := parameterOffset + len(trans2Params)
		pad2 := (4 - (afterParams % 4)) % 4
		cmd.TotalDataCount = types.USHORT(len(trans2Data))
		cmd.DataCount = types.USHORT(len(trans2Data))
		cmd.DataOffset = types.USHORT(afterParams + pad2)
		cmd.Pad2 = make([]types.UCHAR, pad2)
		cmd.Trans2_Data = []types.UCHAR(trans2Data)
	}

	msg.AddCommand(cmd)

	response, raw, err := c.sendReceive(msg, "Transaction2")
	if err != nil {
		return nil, nil, err
	}
	if response.Header.Status != 0x00000000 {
		return nil, nil, fmt.Errorf("Transaction2 (subcommand 0x%04x) failed: 0x%08x", subcommand, response.Header.Status)
	}

	return parseTrans2Response(raw)
}

// parseTrans2Response extracts the transaction parameter and data buffers from a
// raw SMB_COM_TRANSACTION2 response, using the response ParameterOffset/DataOffset
// (which are measured from the start of the SMB header) to locate them.
func parseTrans2Response(raw []byte) ([]byte, []byte, error) {
	hdr := header.SMB_HEADER_SIZE
	if len(raw) < hdr+1 {
		return nil, nil, fmt.Errorf("Transaction2 response too short")
	}

	wordCount := int(raw[hdr])
	wordsStart := hdr + 1
	if len(raw) < wordsStart+2*wordCount {
		return nil, nil, fmt.Errorf("Transaction2 response parameter words truncated")
	}
	// The standard response has WordCount 0x0A (10 words) before any setup words.
	if wordCount < 10 {
		return []byte{}, []byte{}, nil
	}

	w := raw[wordsStart:]
	paramCount := int(binary.LittleEndian.Uint16(w[6:8]))
	paramOffset := int(binary.LittleEndian.Uint16(w[8:10]))
	dataCount := int(binary.LittleEndian.Uint16(w[12:14]))
	dataOffset := int(binary.LittleEndian.Uint16(w[14:16]))

	var params, data []byte
	if paramCount > 0 {
		if paramOffset < 0 || paramOffset+paramCount > len(raw) {
			return nil, nil, fmt.Errorf("Transaction2 parameter window [%d:%d] out of bounds (%d bytes)", paramOffset, paramOffset+paramCount, len(raw))
		}
		params = raw[paramOffset : paramOffset+paramCount]
	}
	if dataCount > 0 {
		if dataOffset < 0 || dataOffset+dataCount > len(raw) {
			return nil, nil, fmt.Errorf("Transaction2 data window [%d:%d] out of bounds (%d bytes)", dataOffset, dataOffset+dataCount, len(raw))
		}
		data = raw[dataOffset : dataOffset+dataCount]
	}

	return params, data, nil
}

// parseBothDirInfo decodes a buffer of consecutive SMB_FIND_FILE_BOTH_DIRECTORY_INFO
// entries (as returned in the FIND_FIRST2/FIND_NEXT2 data) into FileEntry values.
// FileName is decoded as OEM/ASCII because the client issues non-Unicode requests.
func parseBothDirInfo(data []byte) []FileEntry {
	entries := []FileEntry{}

	for pos := 0; pos+bothDirInfoFixedSize <= len(data); {
		next := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
		createdAt := filetimeToTime(binary.LittleEndian.Uint64(data[pos+8 : pos+16]))
		accessedAt := filetimeToTime(binary.LittleEndian.Uint64(data[pos+16 : pos+24]))
		modifiedAt := filetimeToTime(binary.LittleEndian.Uint64(data[pos+24 : pos+32]))
		changedAt := filetimeToTime(binary.LittleEndian.Uint64(data[pos+32 : pos+40]))
		size := binary.LittleEndian.Uint64(data[pos+40 : pos+48])
		attrs := binary.LittleEndian.Uint32(data[pos+56 : pos+60])
		nameLen := int(binary.LittleEndian.Uint32(data[pos+60 : pos+64]))
		shortNameLen := int(data[pos+68])

		shortName := ""
		if shortNameLen > 0 && shortNameLen <= 24 {
			shortName = decodeUTF16LE(data[pos+70 : pos+70+shortNameLen])
		}

		longName := ""
		if nameLen > 0 && pos+bothDirInfoFixedSize+nameLen <= len(data) {
			longName = string(data[pos+bothDirInfoFixedSize : pos+bothDirInfoFixedSize+nameLen])
		}

		entries = append(entries, FileEntry{
			LongName:    longName,
			ShortName:   shortName,
			Size:        size,
			Attributes:  attrs,
			CreatedAt:   createdAt,
			AccessedAt:  accessedAt,
			ModifiedAt:  modifiedAt,
			ChangedAt:   changedAt,
			IsDirectory: attrs&fileAttributeDirectory != 0,
		})

		if next == 0 {
			break
		}
		pos += next
	}

	return entries
}

// findClose2 releases a search handle obtained from FIND_FIRST2.
//
// Wire: FindClose2Request / FindClose2Response.
func (c *Client) findClose2(sid uint16) error {
	msg := c.newFileIOMessage(codes.SMB_COM_FIND_CLOSE2)

	cmd := commands.NewFindClose2Request()
	cmd.SearchHandle = types.USHORT(sid)

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "FindClose2")
	if err != nil {
		return err
	}
	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("FindClose2 failed: 0x%08x", response.Header.Status)
	}
	return nil
}

// buildFindFirst2Params builds the TRANS2_FIND_FIRST2 transaction parameters.
func buildFindFirst2Params(pattern string) []byte {
	b := []byte{}
	b = binary.LittleEndian.AppendUint16(b, 0x0016) // SearchAttributes: include hidden/system/directory
	b = binary.LittleEndian.AppendUint16(b, 512)    // SearchCount: max entries per response
	b = binary.LittleEndian.AppendUint16(b, 0x0000) // Flags: none (the search is closed explicitly)
	b = binary.LittleEndian.AppendUint16(b, smbFindFileBothDirectoryInfo)
	b = binary.LittleEndian.AppendUint32(b, 0) // SearchStorageType
	b = append(b, []byte(pattern)...)
	b = append(b, 0x00) // null-terminated OEM pattern
	return b
}

// buildFindNext2Params builds the TRANS2_FIND_NEXT2 transaction parameters that
// continue a search from the server's cursor.
func buildFindNext2Params(sid uint16) []byte {
	b := []byte{}
	b = binary.LittleEndian.AppendUint16(b, sid) // SID (search handle)
	b = binary.LittleEndian.AppendUint16(b, 512) // SearchCount
	b = binary.LittleEndian.AppendUint16(b, smbFindFileBothDirectoryInfo)
	b = binary.LittleEndian.AppendUint32(b, 0)      // ResumeKey (unused; we did not request resume keys)
	b = binary.LittleEndian.AppendUint16(b, 0x0008) // Flags: SMB_FIND_CONTINUE_FROM_LAST
	b = append(b, 0x00)                             // FileName: empty (resume from server cursor)
	return b
}

// filetimeToTime converts a Windows FILETIME (100ns ticks since 1601-01-01 UTC)
// to a time.Time. A zero FILETIME maps to the zero time.
func filetimeToTime(ft uint64) time.Time {
	const epochDiff = 116444736000000000 // 100ns intervals between 1601-01-01 and 1970-01-01
	if ft == 0 || ft < epochDiff {
		return time.Time{}
	}
	return time.Unix(0, int64(ft-epochDiff)*100).UTC()
}

// decodeUTF16LE decodes a little-endian UTF-16 byte buffer into a string, trimming
// any trailing NUL units.
func decodeUTF16LE(b []byte) string {
	u16 := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		u16 = append(u16, binary.LittleEndian.Uint16(b[i:i+2]))
	}
	return strings.TrimRight(string(utf16.Decode(u16)), "\x00")
}
