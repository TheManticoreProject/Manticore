package ms_fscc

import (
	"encoding/binary"
	"fmt"
)

// FileBasicInformation is FILE_BASIC_INFORMATION (FileInformationClass 4): the
// timestamps and attributes of a file. FILETIME fields are 100ns ticks since
// 1601-01-01 UTC. Fixed 40-byte wire layout.
//
// [MS-FSCC] 2.4.7 FileBasicInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/16023025-8a78-492f-8b96-c873b042ac50
type FileBasicInformation struct {
	CreationTime   uint64
	LastAccessTime uint64
	LastWriteTime  uint64
	ChangeTime     uint64
	FileAttributes uint32
	Reserved       uint32
}

// FileBasicInformationSize is the fixed wire size of FILE_BASIC_INFORMATION.
const FileBasicInformationSize = 40

// Marshal encodes the structure to its 40-byte wire form.
func (fi *FileBasicInformation) Marshal() ([]byte, error) {
	b := make([]byte, FileBasicInformationSize)
	binary.LittleEndian.PutUint64(b[0:8], fi.CreationTime)
	binary.LittleEndian.PutUint64(b[8:16], fi.LastAccessTime)
	binary.LittleEndian.PutUint64(b[16:24], fi.LastWriteTime)
	binary.LittleEndian.PutUint64(b[24:32], fi.ChangeTime)
	binary.LittleEndian.PutUint32(b[32:36], fi.FileAttributes)
	binary.LittleEndian.PutUint32(b[36:40], fi.Reserved)
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileBasicInformation) Unmarshal(data []byte) error {
	if len(data) < FileBasicInformationSize {
		return fmt.Errorf("FILE_BASIC_INFORMATION: need %d bytes, got %d", FileBasicInformationSize, len(data))
	}
	fi.CreationTime = binary.LittleEndian.Uint64(data[0:8])
	fi.LastAccessTime = binary.LittleEndian.Uint64(data[8:16])
	fi.LastWriteTime = binary.LittleEndian.Uint64(data[16:24])
	fi.ChangeTime = binary.LittleEndian.Uint64(data[24:32])
	fi.FileAttributes = binary.LittleEndian.Uint32(data[32:36])
	fi.Reserved = binary.LittleEndian.Uint32(data[36:40])
	return nil
}

// FileStandardInformation is FILE_STANDARD_INFORMATION (FileInformationClass 5):
// size, link count, and delete/directory flags. Fixed 24-byte wire layout.
//
// [MS-FSCC] 2.4.41 FileStandardInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/5afa7f66-619c-48f3-955f-68c4ece704ae
type FileStandardInformation struct {
	AllocationSize int64
	EndOfFile      int64
	NumberOfLinks  uint32
	DeletePending  bool
	Directory      bool
}

// FileStandardInformationSize is the fixed wire size of FILE_STANDARD_INFORMATION.
const FileStandardInformationSize = 24

// Unmarshal decodes the structure from its wire form.
func (fi *FileStandardInformation) Unmarshal(data []byte) error {
	if len(data) < FileStandardInformationSize {
		return fmt.Errorf("FILE_STANDARD_INFORMATION: need %d bytes, got %d", FileStandardInformationSize, len(data))
	}
	fi.AllocationSize = int64(binary.LittleEndian.Uint64(data[0:8]))
	fi.EndOfFile = int64(binary.LittleEndian.Uint64(data[8:16]))
	fi.NumberOfLinks = binary.LittleEndian.Uint32(data[16:20])
	fi.DeletePending = data[20] != 0
	fi.Directory = data[21] != 0
	return nil
}

// Marshal encodes the structure to its 24-byte wire form.
func (fi *FileStandardInformation) Marshal() ([]byte, error) {
	b := make([]byte, FileStandardInformationSize)
	binary.LittleEndian.PutUint64(b[0:8], uint64(fi.AllocationSize))
	binary.LittleEndian.PutUint64(b[8:16], uint64(fi.EndOfFile))
	binary.LittleEndian.PutUint32(b[16:20], fi.NumberOfLinks)
	if fi.DeletePending {
		b[20] = 1
	}
	if fi.Directory {
		b[21] = 1
	}
	return b, nil
}
