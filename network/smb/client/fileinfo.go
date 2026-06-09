package client

import (
	"encoding/binary"
	"strings"
	"time"
	"unicode/utf16"
)

// fileAttributeDirectory is the FILE_ATTRIBUTE_DIRECTORY bit (MS-FSCC 2.6).
const fileAttributeDirectory uint32 = 0x00000010

// File-open and information-class constants used to enumerate a directory over
// SMB2 (MS-SMB2 / MS-FSCC).
const (
	fileListDirectory  uint32 = 0x00000001 // DesiredAccess: FILE_LIST_DIRECTORY
	fileReadAttributes uint32 = 0x00000080 // DesiredAccess: FILE_READ_ATTRIBUTES
	fileOpen           uint32 = 0x00000001 // CreateDisposition: FILE_OPEN
	fileDirectoryFile  uint32 = 0x00000001 // CreateOptions: FILE_DIRECTORY_FILE
	shareReadWrite     uint32 = 0x00000003 // ShareAccess: FILE_SHARE_READ|WRITE

	// fileBothDirectoryInformation is the FILE_INFORMATION_CLASS used for
	// enumeration; its FILE_BOTH_DIR_INFORMATION entries (MS-FSCC 2.4.8) carry the
	// name, attributes, size, and timestamps needed for FileInfo.
	fileBothDirectoryInformation uint8 = 0x03
)

// bothDirInfoFixedSize is the fixed portion of a FILE_BOTH_DIR_INFORMATION entry
// before the variable FileName: NextEntryOffset(4) FileIndex(4) 4xFILETIME(32)
// EndOfFile(8) AllocationSize(8) FileAttributes(4) FileNameLength(4) EaSize(4)
// ShortNameLength(1) Reserved(1) ShortName(24).
const bothDirInfoFixedSize = 94

// parseBothDirectoryInfo decodes a buffer of consecutive FILE_BOTH_DIR_INFORMATION
// entries (MS-FSCC 2.4.8), as returned by SMB2 QUERY_DIRECTORY, into FileInfo
// values. FileName is UTF-16LE. The "." and ".." pseudo-entries are skipped.
//
// A typed MS-FSCC information-class package (#525) would replace this hand-rolled
// decoder; it lives here until that lands.
func parseBothDirectoryInfo(buf []byte) []FileInfo {
	var out []FileInfo

	for pos := 0; pos+bothDirInfoFixedSize <= len(buf); {
		next := int(binary.LittleEndian.Uint32(buf[pos : pos+4]))
		creation := filetimeToTime(binary.LittleEndian.Uint64(buf[pos+8 : pos+16]))
		lastAccess := filetimeToTime(binary.LittleEndian.Uint64(buf[pos+16 : pos+24]))
		lastWrite := filetimeToTime(binary.LittleEndian.Uint64(buf[pos+24 : pos+32]))
		change := filetimeToTime(binary.LittleEndian.Uint64(buf[pos+32 : pos+40]))
		endOfFile := binary.LittleEndian.Uint64(buf[pos+40 : pos+48])
		allocationSize := binary.LittleEndian.Uint64(buf[pos+48 : pos+56])
		attrs := binary.LittleEndian.Uint32(buf[pos+56 : pos+60])
		nameLen := int(binary.LittleEndian.Uint32(buf[pos+60 : pos+64]))

		name := ""
		if nameLen > 0 && pos+bothDirInfoFixedSize+nameLen <= len(buf) {
			name = decodeUTF16LE(buf[pos+bothDirInfoFixedSize : pos+bothDirInfoFixedSize+nameLen])
		}

		if name != "." && name != ".." {
			out = append(out, FileInfo{
				Name:           name,
				FileAttributes: attrs,
				Size:           endOfFile,
				AllocationSize: allocationSize,
				CreationTime:   creation,
				LastAccessTime: lastAccess,
				LastWriteTime:  lastWrite,
				ChangeTime:     change,
			})
		}

		if next == 0 {
			break
		}
		pos += next
	}

	return out
}

// filetimeToTime converts a Windows FILETIME (100ns ticks since 1601-01-01 UTC)
// to a time.Time. A zero (or sub-epoch) FILETIME maps to the zero time.
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
