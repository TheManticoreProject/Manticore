package filesystem

import (
	"encoding/binary"
	"unicode/utf16"
)

// CompletionFilter flags for a directory change notification: the set of changes
// to report (SMB2 CHANGE_NOTIFY CompletionFilter / FILE_NOTIFY_CHANGE_*).
//
// [MS-SMB2] 2.2.35 SMB2 CHANGE_NOTIFY Request:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/598f395a-e7a2-4cc8-afb3-ccb30dd2df7c
const (
	FILE_NOTIFY_CHANGE_FILE_NAME    uint32 = 0x00000001
	FILE_NOTIFY_CHANGE_DIR_NAME     uint32 = 0x00000002
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

// FileAction is the change reported for a directory entry in a
// FILE_NOTIFY_INFORMATION record.
type FileAction uint32

const (
	FileActionAdded          FileAction = 0x00000001
	FileActionRemoved        FileAction = 0x00000002
	FileActionModified       FileAction = 0x00000003
	FileActionRenamedOldName FileAction = 0x00000004
	FileActionRenamedNewName FileAction = 0x00000005
)

// String renders the action name.
func (a FileAction) String() string {
	switch a {
	case FileActionAdded:
		return "ADDED"
	case FileActionRemoved:
		return "REMOVED"
	case FileActionModified:
		return "MODIFIED"
	case FileActionRenamedOldName:
		return "RENAMED_OLD_NAME"
	case FileActionRenamedNewName:
		return "RENAMED_NEW_NAME"
	}
	return "UNKNOWN"
}

// FileNotifyInformation is one FILE_NOTIFY_INFORMATION record: a change to a named
// entry within a watched directory.
//
// [MS-FSCC] 2.7.1 FILE_NOTIFY_INFORMATION:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/634043d7-7b39-47e9-9e26-bda64685e4c9
type FileNotifyInformation struct {
	Action   FileAction
	FileName string // directory-relative, backslash-separated
}

// ParseFileNotifyInformation decodes a buffer of chained FILE_NOTIFY_INFORMATION
// records (as returned by SMB2 CHANGE_NOTIFY) into FileNotifyInformation values.
// FileName is UTF-16LE.
func ParseFileNotifyInformation(buf []byte) []FileNotifyInformation {
	var out []FileNotifyInformation
	for pos := 0; pos+12 <= len(buf); {
		next := int(binary.LittleEndian.Uint32(buf[pos : pos+4]))
		action := FileAction(binary.LittleEndian.Uint32(buf[pos+4 : pos+8]))
		nameLen := int(binary.LittleEndian.Uint32(buf[pos+8 : pos+12]))

		name := ""
		if nameLen > 0 && pos+12+nameLen <= len(buf) {
			u16 := make([]uint16, nameLen/2)
			for i := range u16 {
				u16[i] = binary.LittleEndian.Uint16(buf[pos+12+i*2:])
			}
			name = string(utf16.Decode(u16))
		}
		out = append(out, FileNotifyInformation{Action: action, FileName: name})

		if next == 0 {
			break
		}
		pos += next
	}
	return out
}
