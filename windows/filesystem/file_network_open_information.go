package filesystem

import (
	"encoding/binary"
	"fmt"
)

// FileNetworkOpenInformation is FILE_NETWORK_OPEN_INFORMATION
// (FileInformationClass 34): the timestamps, sizes, and attributes returned for a
// network-open query, equivalent to the data a CREATE response carries. FILETIME
// fields are 100ns ticks since 1601-01-01 UTC. Fixed 56-byte wire layout.
//
// [MS-FSCC] 2.4.29 FileNetworkOpenInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/b8e3957e-f16e-49b8-b8b1-58b9b22f3f54
type FileNetworkOpenInformation struct {
	CreationTime   uint64
	LastAccessTime uint64
	LastWriteTime  uint64
	ChangeTime     uint64
	AllocationSize int64
	EndOfFile      int64
	FileAttributes uint32
	Reserved       uint32
}

// FileNetworkOpenInformationSize is the fixed wire size of
// FILE_NETWORK_OPEN_INFORMATION.
const FileNetworkOpenInformationSize = 56

// Marshal encodes the structure to its 56-byte wire form.
func (fi *FileNetworkOpenInformation) Marshal() ([]byte, error) {
	b := make([]byte, FileNetworkOpenInformationSize)
	binary.LittleEndian.PutUint64(b[0:8], fi.CreationTime)
	binary.LittleEndian.PutUint64(b[8:16], fi.LastAccessTime)
	binary.LittleEndian.PutUint64(b[16:24], fi.LastWriteTime)
	binary.LittleEndian.PutUint64(b[24:32], fi.ChangeTime)
	binary.LittleEndian.PutUint64(b[32:40], uint64(fi.AllocationSize))
	binary.LittleEndian.PutUint64(b[40:48], uint64(fi.EndOfFile))
	binary.LittleEndian.PutUint32(b[48:52], fi.FileAttributes)
	binary.LittleEndian.PutUint32(b[52:56], fi.Reserved)
	return b, nil
}

// Unmarshal decodes the structure from its wire form.
func (fi *FileNetworkOpenInformation) Unmarshal(data []byte) error {
	if len(data) < FileNetworkOpenInformationSize {
		return fmt.Errorf("FILE_NETWORK_OPEN_INFORMATION: need %d bytes, got %d", FileNetworkOpenInformationSize, len(data))
	}
	fi.CreationTime = binary.LittleEndian.Uint64(data[0:8])
	fi.LastAccessTime = binary.LittleEndian.Uint64(data[8:16])
	fi.LastWriteTime = binary.LittleEndian.Uint64(data[16:24])
	fi.ChangeTime = binary.LittleEndian.Uint64(data[24:32])
	fi.AllocationSize = int64(binary.LittleEndian.Uint64(data[32:40]))
	fi.EndOfFile = int64(binary.LittleEndian.Uint64(data[40:48]))
	fi.FileAttributes = binary.LittleEndian.Uint32(data[48:52])
	fi.Reserved = binary.LittleEndian.Uint32(data[52:56])
	return nil
}
