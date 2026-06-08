package subcommands

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// NT_TRANSACT_NOTIFY_CHANGE ([MS-CIFS] section 2.2.7.4) lets a client watch an open
// directory for changes. The request carries its arguments in the NT_Trans setup words
// (no NT_Trans_Parameters/Data); the response returns the changed-object records as a
// list of FILE_NOTIFY_INFORMATION structures in the NT_Trans_Parameters block.

// CompletionFilter flags for the NT_TRANSACT_NOTIFY_CHANGE setup ([MS-CIFS] section
// 2.2.7.4.1): the set of change events the client wants to be notified about.
const (
	FILE_NOTIFY_CHANGE_FILE_NAME    uint32 = 0x00000001
	FILE_NOTIFY_CHANGE_DIR_NAME     uint32 = 0x00000002
	FILE_NOTIFY_CHANGE_NAME         uint32 = 0x00000003
	FILE_NOTIFY_CHANGE_ATTRIBUTES   uint32 = 0x00000004
	FILE_NOTIFY_CHANGE_SIZE         uint32 = 0x00000008
	FILE_NOTIFY_CHANGE_LAST_WRITE   uint32 = 0x00000010
	FILE_NOTIFY_CHANGE_LAST_ACCESS  uint32 = 0x00000020
	FILE_NOTIFY_CHANGE_CREATION     uint32 = 0x00000040
	FILE_NOTIFY_CHANGE_EA           uint32 = 0x00000080
	FILE_NOTIFY_CHANGE_SECURITY     uint32 = 0x00000100
	FILE_NOTIFY_CHANGE_STREAM_NAME  uint32 = 0x00000200
	FILE_NOTIFY_CHANGE_STREAM_SIZE  uint32 = 0x00000400
	FILE_NOTIFY_CHANGE_STREAM_WRITE uint32 = 0x00000800
)

// FILE_ACTION_* values for FileNotifyInformation.Action ([MS-FSCC] section 2.7.1): the
// kind of change that occurred to the named object.
const (
	FILE_ACTION_ADDED            uint32 = 0x00000001
	FILE_ACTION_REMOVED          uint32 = 0x00000002
	FILE_ACTION_MODIFIED         uint32 = 0x00000003
	FILE_ACTION_RENAMED_OLD_NAME uint32 = 0x00000004
	FILE_ACTION_RENAMED_NEW_NAME uint32 = 0x00000005
)

const (
	ntTransactNotifyChangeSetupSize = 8  // CompletionFilter(4) + FID(2) + WatchTree(1) + Reserved(1)
	fileNotifyInformationHeaderSize = 12 // NextEntryOffset(4) + Action(4) + FileNameLength(4)
)

// utf16leBytes encodes a Go string as little-endian UTF-16 octets.
func utf16leBytes(s string) []byte {
	units := utf16.Encode([]rune(s))
	b := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(b[i*2:], u)
	}
	return b
}

// decodeUTF16LE decodes little-endian UTF-16 octets to a Go string (a trailing odd byte,
// if any, is ignored).
func decodeUTF16LE(b []byte) string {
	units := make([]uint16, len(b)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return string(utf16.Decode(units))
}

// NtTransactNotifyChangeSetup is the NT_Trans setup of an NT_TRANSACT_NOTIFY_CHANGE
// request ([MS-CIFS] section 2.2.7.4.1). The request carries no NT_Trans_Parameters or
// NT_Trans_Data.
type NtTransactNotifyChangeSetup struct {
	// CompletionFilter (4 bytes): the change events to monitor (FILE_NOTIFY_CHANGE_* flags).
	CompletionFilter uint32
	// FID (2 bytes): the open directory to monitor.
	FID uint16
	// WatchTree (1 byte): if true, all subdirectories below FID are also watched.
	WatchTree bool
	// Reserved (1 byte): MUST be 0x00.
	Reserved uint8
}

// Marshal serializes the 8-octet NT_Trans setup.
func (s *NtTransactNotifyChangeSetup) Marshal() ([]byte, error) {
	b := make([]byte, ntTransactNotifyChangeSetupSize)
	binary.LittleEndian.PutUint32(b[0:4], s.CompletionFilter)
	binary.LittleEndian.PutUint16(b[4:6], s.FID)
	b[6] = boolToByte(s.WatchTree)
	b[7] = s.Reserved
	return b, nil
}

// Unmarshal parses the 8-octet NT_Trans setup.
func (s *NtTransactNotifyChangeSetup) Unmarshal(data []byte) (int, error) {
	if len(data) < ntTransactNotifyChangeSetupSize {
		return 0, fmt.Errorf("subcommands: NT_TRANSACT_NOTIFY_CHANGE setup requires %d bytes, got %d", ntTransactNotifyChangeSetupSize, len(data))
	}
	s.CompletionFilter = binary.LittleEndian.Uint32(data[0:4])
	s.FID = binary.LittleEndian.Uint16(data[4:6])
	s.WatchTree = data[6] != 0
	s.Reserved = data[7]
	return ntTransactNotifyChangeSetupSize, nil
}

// FileNotifyInformation is one changed-object record returned in the
// NT_TRANSACT_NOTIFY_CHANGE response ([MS-FSCC] section 2.7.1). FileNameLength (in octets)
// is derived from FileName on marshal.
type FileNotifyInformation struct {
	// NextEntryOffset (4 bytes): octet offset from the start of this record to the next, or
	// zero for the last record. Always a multiple of 4. Set by MarshalFileNotifyInformationList.
	NextEntryOffset uint32
	// Action (4 bytes): the change that occurred (FILE_ACTION_* value).
	Action uint32
	// FileName: the changed object's name, relative to the watched directory (UTF-16LE on
	// the wire, not NUL-terminated).
	FileName string
}

// Marshal serializes a single FILE_NOTIFY_INFORMATION record (NextEntryOffset is emitted
// from the field as-is; use MarshalFileNotifyInformationList to chain a list).
func (f *FileNotifyInformation) Marshal() ([]byte, error) {
	nameBytes := utf16leBytes(f.FileName)
	b := make([]byte, fileNotifyInformationHeaderSize+len(nameBytes))
	binary.LittleEndian.PutUint32(b[0:4], f.NextEntryOffset)
	binary.LittleEndian.PutUint32(b[4:8], f.Action)
	binary.LittleEndian.PutUint32(b[8:12], uint32(len(nameBytes)))
	copy(b[fileNotifyInformationHeaderSize:], nameBytes)
	return b, nil
}

// Unmarshal parses a single FILE_NOTIFY_INFORMATION record, returning the number of octets
// the record itself occupies (header + FileName, ignoring any trailing alignment padding).
func (f *FileNotifyInformation) Unmarshal(data []byte) (int, error) {
	if len(data) < fileNotifyInformationHeaderSize {
		return 0, fmt.Errorf("subcommands: FILE_NOTIFY_INFORMATION requires at least %d bytes, got %d", fileNotifyInformationHeaderSize, len(data))
	}
	f.NextEntryOffset = binary.LittleEndian.Uint32(data[0:4])
	f.Action = binary.LittleEndian.Uint32(data[4:8])
	nameLength := int(binary.LittleEndian.Uint32(data[8:12]))
	if len(data) < fileNotifyInformationHeaderSize+nameLength {
		return fileNotifyInformationHeaderSize, fmt.Errorf("subcommands: FILE_NOTIFY_INFORMATION FileName truncated: need %d, have %d", nameLength, len(data)-fileNotifyInformationHeaderSize)
	}
	f.FileName = decodeUTF16LE(data[fileNotifyInformationHeaderSize : fileNotifyInformationHeaderSize+nameLength])
	return fileNotifyInformationHeaderSize + nameLength, nil
}

// align4 rounds n up to the next multiple of 4.
func align4(n int) int { return (n + 3) &^ 3 }

// MarshalFileNotifyInformationList serializes the NT_TRANSACT_NOTIFY_CHANGE response
// NT_Trans_Parameters: a chained list of FILE_NOTIFY_INFORMATION records. Each record's
// NextEntryOffset is computed (4-byte aligned) so the list is correctly linked, with the
// final record's NextEntryOffset set to zero.
func MarshalFileNotifyInformationList(items []FileNotifyInformation) ([]byte, error) {
	out := []byte{}
	for i := range items {
		nameBytes := utf16leBytes(items[i].FileName)
		recordLen := fileNotifyInformationHeaderSize + len(nameBytes)
		padded := align4(recordLen)

		var next uint32
		if i < len(items)-1 {
			next = uint32(padded)
		}

		hdr := make([]byte, fileNotifyInformationHeaderSize)
		binary.LittleEndian.PutUint32(hdr[0:4], next)
		binary.LittleEndian.PutUint32(hdr[4:8], items[i].Action)
		binary.LittleEndian.PutUint32(hdr[8:12], uint32(len(nameBytes)))
		out = append(out, hdr...)
		out = append(out, nameBytes...)
		for p := recordLen; p < padded; p++ {
			out = append(out, 0) // alignment padding
		}
	}
	return out, nil
}

// ParseFileNotifyInformationList parses the NT_TRANSACT_NOTIFY_CHANGE response
// NT_Trans_Parameters into its FILE_NOTIFY_INFORMATION records, following NextEntryOffset
// until it is zero or the buffer is exhausted.
func ParseFileNotifyInformationList(data []byte) ([]FileNotifyInformation, error) {
	items := []FileNotifyInformation{}
	offset := 0
	for offset+fileNotifyInformationHeaderSize <= len(data) {
		var rec FileNotifyInformation
		if _, err := rec.Unmarshal(data[offset:]); err != nil {
			return items, err
		}
		items = append(items, rec)
		if rec.NextEntryOffset == 0 {
			break
		}
		offset += int(rec.NextEntryOffset)
	}
	return items, nil
}
