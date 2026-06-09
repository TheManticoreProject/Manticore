package filesystem

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// FileEndOfFileInformation is FILE_END_OF_FILE_INFORMATION (FileInformationClass
// 20): sets the logical end-of-file (size) of a file.
//
// [MS-FSCC] 2.4.13 FileEndOfFileInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/75241cca-3167-472f-8058-a52d77c6bb17
type FileEndOfFileInformation struct {
	EndOfFile int64
}

// Marshal encodes the structure to its 8-byte wire form.
func (fi *FileEndOfFileInformation) Marshal() ([]byte, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(fi.EndOfFile))
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileEndOfFileInformation) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("FILE_END_OF_FILE_INFORMATION: need 8 bytes, got %d", len(data))
	}
	fi.EndOfFile = int64(binary.LittleEndian.Uint64(data[0:8]))
	return nil
}

// FileDispositionInformation is FILE_DISPOSITION_INFORMATION (FileInformationClass
// 13): marks a file for deletion when its last handle is closed.
//
// [MS-FSCC] 2.4.11 FileDispositionInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/12c3dd1c-14f6-4229-9d29-75fb2cb392f6
type FileDispositionInformation struct {
	DeletePending bool
}

// Marshal encodes the structure to its 1-byte wire form.
func (fi *FileDispositionInformation) Marshal() ([]byte, error) {
	b := make([]byte, 1)
	if fi.DeletePending {
		b[0] = 1
	}
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileDispositionInformation) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("FILE_DISPOSITION_INFORMATION: need 1 byte, got %d", len(data))
	}
	fi.DeletePending = data[0] != 0
	return nil
}

// FileRenameInformation is the SMB2 form of FILE_RENAME_INFORMATION
// (FileInformationClass 10): renames or moves a file. FileName is the new
// (share-relative) name; RootDirectory MUST be 0 for network operations.
//
// [MS-FSCC] 2.4.34.2 FileRenameInformation for SMB2:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/52aa0b70-8094-4971-862d-79793f41e6a8
type FileRenameInformation struct {
	ReplaceIfExists bool
	RootDirectory   uint64
	FileName        string
}

// Marshal encodes the structure to its wire form: ReplaceIfExists(1) Reserved(7)
// RootDirectory(8) FileNameLength(4) FileName(UTF-16LE, not NUL-terminated).
func (fi *FileRenameInformation) Marshal() ([]byte, error) {
	name := utf16.Encode([]rune(fi.FileName))
	nameBytes := make([]byte, len(name)*2)
	for i, u := range name {
		binary.LittleEndian.PutUint16(nameBytes[i*2:], u)
	}

	b := make([]byte, 20+len(nameBytes))
	if fi.ReplaceIfExists {
		b[0] = 1
	}
	// b[1:8] Reserved = 0
	binary.LittleEndian.PutUint64(b[8:16], fi.RootDirectory)
	binary.LittleEndian.PutUint32(b[16:20], uint32(len(nameBytes)))
	copy(b[20:], nameBytes)
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileRenameInformation) Unmarshal(data []byte) error {
	if len(data) < 20 {
		return fmt.Errorf("FILE_RENAME_INFORMATION: need at least 20 bytes, got %d", len(data))
	}
	fi.ReplaceIfExists = data[0] != 0
	fi.RootDirectory = binary.LittleEndian.Uint64(data[8:16])
	nameLen := int(binary.LittleEndian.Uint32(data[16:20]))
	if 20+nameLen > len(data) {
		return fmt.Errorf("FILE_RENAME_INFORMATION: name length %d exceeds buffer", nameLen)
	}
	u16 := make([]uint16, nameLen/2)
	for i := range u16 {
		u16[i] = binary.LittleEndian.Uint16(data[20+i*2:])
	}
	fi.FileName = string(utf16.Decode(u16))
	return nil
}
