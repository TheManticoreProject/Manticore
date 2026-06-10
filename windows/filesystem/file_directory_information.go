package filesystem

import (
	"encoding/binary"
	"fmt"
)

// Directory-enumeration information classes (the payloads of SMB2 QUERY_DIRECTORY,
// MS-FSCC 2.4). Each class is a sequence of variable-length entries chained by a
// NextEntryOffset field; the last entry has NextEntryOffset == 0. The per-class
// Unmarshal decodes a single entry from the start of a buffer; the package
// Parse*List helpers walk a full multi-entry buffer.

// FileFullDirectoryInformation is FILE_FULL_DIR_INFORMATION
// (FileInformationClass 2): a directory entry with timestamps, sizes,
// attributes, and EA size.
//
// [MS-FSCC] 2.4.14 FileFullDirectoryInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/e6d24f0e-fa87-4d8d-aa9d-a30c5cf75c0e
type FileFullDirectoryInformation struct {
	NextEntryOffset uint32
	FileIndex       uint32
	CreationTime    uint64
	LastAccessTime  uint64
	LastWriteTime   uint64
	ChangeTime      uint64
	EndOfFile       int64
	AllocationSize  int64
	FileAttributes  uint32
	EaSize          uint32
	FileName        string
}

// fileFullDirInfoFixedSize is the fixed portion preceding FileName.
const fileFullDirInfoFixedSize = 68

// Marshal encodes a single entry (fixed prefix + FileName). NextEntryOffset is
// written verbatim; callers chaining entries set it themselves.
func (fi *FileFullDirectoryInformation) Marshal() ([]byte, error) {
	name := encodeUTF16LE(fi.FileName)
	b := make([]byte, fileFullDirInfoFixedSize+len(name))
	binary.LittleEndian.PutUint32(b[0:4], fi.NextEntryOffset)
	binary.LittleEndian.PutUint32(b[4:8], fi.FileIndex)
	binary.LittleEndian.PutUint64(b[8:16], fi.CreationTime)
	binary.LittleEndian.PutUint64(b[16:24], fi.LastAccessTime)
	binary.LittleEndian.PutUint64(b[24:32], fi.LastWriteTime)
	binary.LittleEndian.PutUint64(b[32:40], fi.ChangeTime)
	binary.LittleEndian.PutUint64(b[40:48], uint64(fi.EndOfFile))
	binary.LittleEndian.PutUint64(b[48:56], uint64(fi.AllocationSize))
	binary.LittleEndian.PutUint32(b[56:60], fi.FileAttributes)
	binary.LittleEndian.PutUint32(b[60:64], uint32(len(name)))
	binary.LittleEndian.PutUint32(b[64:68], fi.EaSize)
	copy(b[fileFullDirInfoFixedSize:], name)
	return b, nil
}

// Unmarshal decodes a single entry from the start of data.
func (fi *FileFullDirectoryInformation) Unmarshal(data []byte) error {
	if len(data) < fileFullDirInfoFixedSize {
		return fmt.Errorf("FILE_FULL_DIR_INFORMATION: need %d bytes, got %d", fileFullDirInfoFixedSize, len(data))
	}
	fi.NextEntryOffset = binary.LittleEndian.Uint32(data[0:4])
	fi.FileIndex = binary.LittleEndian.Uint32(data[4:8])
	fi.CreationTime = binary.LittleEndian.Uint64(data[8:16])
	fi.LastAccessTime = binary.LittleEndian.Uint64(data[16:24])
	fi.LastWriteTime = binary.LittleEndian.Uint64(data[24:32])
	fi.ChangeTime = binary.LittleEndian.Uint64(data[32:40])
	fi.EndOfFile = int64(binary.LittleEndian.Uint64(data[40:48]))
	fi.AllocationSize = int64(binary.LittleEndian.Uint64(data[48:56]))
	fi.FileAttributes = binary.LittleEndian.Uint32(data[56:60])
	nameLen := int(binary.LittleEndian.Uint32(data[60:64]))
	fi.EaSize = binary.LittleEndian.Uint32(data[64:68])
	if nameLen > 0 && fileFullDirInfoFixedSize+nameLen <= len(data) {
		fi.FileName = decodeUTF16LE(data[fileFullDirInfoFixedSize : fileFullDirInfoFixedSize+nameLen])
	}
	return nil
}

// ParseFileFullDirectoryInformationList decodes a chained buffer of
// FILE_FULL_DIR_INFORMATION entries.
func ParseFileFullDirectoryInformationList(buf []byte) []FileFullDirectoryInformation {
	var out []FileFullDirectoryInformation
	for pos := 0; pos+fileFullDirInfoFixedSize <= len(buf); {
		var e FileFullDirectoryInformation
		if err := e.Unmarshal(buf[pos:]); err != nil {
			break
		}
		out = append(out, e)
		if e.NextEntryOffset == 0 {
			break
		}
		pos += int(e.NextEntryOffset)
	}
	return out
}

// FileBothDirectoryInformation is FILE_BOTH_DIR_INFORMATION (FileInformationClass
// 3): a FILE_FULL_DIR_INFORMATION entry plus the 8.3 short name.
//
// [MS-FSCC] 2.4.8 FileBothDirectoryInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/270df317-9ba5-4ccb-ba00-8d22be139bc5
type FileBothDirectoryInformation struct {
	NextEntryOffset uint32
	FileIndex       uint32
	CreationTime    uint64
	LastAccessTime  uint64
	LastWriteTime   uint64
	ChangeTime      uint64
	EndOfFile       int64
	AllocationSize  int64
	FileAttributes  uint32
	EaSize          uint32
	ShortName       string // 8.3 name, UTF-16LE, max 12 chars
	FileName        string
}

// fileBothDirInfoFixedSize is the fixed portion preceding FileName: the full-dir
// prefix (68) plus ShortNameLength(1) Reserved(1) ShortName(24).
const fileBothDirInfoFixedSize = 94

// Marshal encodes a single entry (fixed prefix + FileName).
func (fi *FileBothDirectoryInformation) Marshal() ([]byte, error) {
	name := encodeUTF16LE(fi.FileName)
	short := encodeUTF16LE(fi.ShortName)
	if len(short) > 24 {
		return nil, fmt.Errorf("FILE_BOTH_DIR_INFORMATION: short name exceeds 24 bytes (got %d)", len(short))
	}
	b := make([]byte, fileBothDirInfoFixedSize+len(name))
	binary.LittleEndian.PutUint32(b[0:4], fi.NextEntryOffset)
	binary.LittleEndian.PutUint32(b[4:8], fi.FileIndex)
	binary.LittleEndian.PutUint64(b[8:16], fi.CreationTime)
	binary.LittleEndian.PutUint64(b[16:24], fi.LastAccessTime)
	binary.LittleEndian.PutUint64(b[24:32], fi.LastWriteTime)
	binary.LittleEndian.PutUint64(b[32:40], fi.ChangeTime)
	binary.LittleEndian.PutUint64(b[40:48], uint64(fi.EndOfFile))
	binary.LittleEndian.PutUint64(b[48:56], uint64(fi.AllocationSize))
	binary.LittleEndian.PutUint32(b[56:60], fi.FileAttributes)
	binary.LittleEndian.PutUint32(b[60:64], uint32(len(name)))
	binary.LittleEndian.PutUint32(b[64:68], fi.EaSize)
	b[68] = byte(len(short))
	// b[69] Reserved = 0
	copy(b[70:94], short)
	copy(b[fileBothDirInfoFixedSize:], name)
	return b, nil
}

// Unmarshal decodes a single entry from the start of data.
func (fi *FileBothDirectoryInformation) Unmarshal(data []byte) error {
	if len(data) < fileBothDirInfoFixedSize {
		return fmt.Errorf("FILE_BOTH_DIR_INFORMATION: need %d bytes, got %d", fileBothDirInfoFixedSize, len(data))
	}
	fi.NextEntryOffset = binary.LittleEndian.Uint32(data[0:4])
	fi.FileIndex = binary.LittleEndian.Uint32(data[4:8])
	fi.CreationTime = binary.LittleEndian.Uint64(data[8:16])
	fi.LastAccessTime = binary.LittleEndian.Uint64(data[16:24])
	fi.LastWriteTime = binary.LittleEndian.Uint64(data[24:32])
	fi.ChangeTime = binary.LittleEndian.Uint64(data[32:40])
	fi.EndOfFile = int64(binary.LittleEndian.Uint64(data[40:48]))
	fi.AllocationSize = int64(binary.LittleEndian.Uint64(data[48:56]))
	fi.FileAttributes = binary.LittleEndian.Uint32(data[56:60])
	nameLen := int(binary.LittleEndian.Uint32(data[60:64]))
	fi.EaSize = binary.LittleEndian.Uint32(data[64:68])
	shortLen := int(data[68])
	if shortLen > 24 {
		shortLen = 24
	}
	fi.ShortName = decodeUTF16LE(data[70 : 70+shortLen])
	if nameLen > 0 && fileBothDirInfoFixedSize+nameLen <= len(data) {
		fi.FileName = decodeUTF16LE(data[fileBothDirInfoFixedSize : fileBothDirInfoFixedSize+nameLen])
	}
	return nil
}

// ParseFileBothDirectoryInformationList decodes a chained buffer of
// FILE_BOTH_DIR_INFORMATION entries.
func ParseFileBothDirectoryInformationList(buf []byte) []FileBothDirectoryInformation {
	var out []FileBothDirectoryInformation
	for pos := 0; pos+fileBothDirInfoFixedSize <= len(buf); {
		var e FileBothDirectoryInformation
		if err := e.Unmarshal(buf[pos:]); err != nil {
			break
		}
		out = append(out, e)
		if e.NextEntryOffset == 0 {
			break
		}
		pos += int(e.NextEntryOffset)
	}
	return out
}

// FileIdBothDirectoryInformation is FILE_ID_BOTH_DIR_INFORMATION
// (FileInformationClass 37): a FILE_BOTH_DIR_INFORMATION entry plus the 64-bit
// file id.
//
// [MS-FSCC] 2.4.17 FileIdBothDirectoryInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/9b0b9971-85aa-4651-8438-f1c4298bcb0d
type FileIdBothDirectoryInformation struct {
	NextEntryOffset uint32
	FileIndex       uint32
	CreationTime    uint64
	LastAccessTime  uint64
	LastWriteTime   uint64
	ChangeTime      uint64
	EndOfFile       int64
	AllocationSize  int64
	FileAttributes  uint32
	EaSize          uint32
	ShortName       string
	FileId          uint64
	FileName        string
}

// fileIdBothDirInfoFixedSize is the fixed portion preceding FileName: the both-dir
// prefix (94) plus Reserved2(2) FileId(8).
const fileIdBothDirInfoFixedSize = 104

// Marshal encodes a single entry (fixed prefix + FileName).
func (fi *FileIdBothDirectoryInformation) Marshal() ([]byte, error) {
	name := encodeUTF16LE(fi.FileName)
	short := encodeUTF16LE(fi.ShortName)
	if len(short) > 24 {
		return nil, fmt.Errorf("FILE_ID_BOTH_DIR_INFORMATION: short name exceeds 24 bytes (got %d)", len(short))
	}
	b := make([]byte, fileIdBothDirInfoFixedSize+len(name))
	binary.LittleEndian.PutUint32(b[0:4], fi.NextEntryOffset)
	binary.LittleEndian.PutUint32(b[4:8], fi.FileIndex)
	binary.LittleEndian.PutUint64(b[8:16], fi.CreationTime)
	binary.LittleEndian.PutUint64(b[16:24], fi.LastAccessTime)
	binary.LittleEndian.PutUint64(b[24:32], fi.LastWriteTime)
	binary.LittleEndian.PutUint64(b[32:40], fi.ChangeTime)
	binary.LittleEndian.PutUint64(b[40:48], uint64(fi.EndOfFile))
	binary.LittleEndian.PutUint64(b[48:56], uint64(fi.AllocationSize))
	binary.LittleEndian.PutUint32(b[56:60], fi.FileAttributes)
	binary.LittleEndian.PutUint32(b[60:64], uint32(len(name)))
	binary.LittleEndian.PutUint32(b[64:68], fi.EaSize)
	b[68] = byte(len(short))
	// b[69] Reserved1 = 0
	copy(b[70:94], short)
	// b[94:96] Reserved2 = 0
	binary.LittleEndian.PutUint64(b[96:104], fi.FileId)
	copy(b[fileIdBothDirInfoFixedSize:], name)
	return b, nil
}

// Unmarshal decodes a single entry from the start of data.
func (fi *FileIdBothDirectoryInformation) Unmarshal(data []byte) error {
	if len(data) < fileIdBothDirInfoFixedSize {
		return fmt.Errorf("FILE_ID_BOTH_DIR_INFORMATION: need %d bytes, got %d", fileIdBothDirInfoFixedSize, len(data))
	}
	fi.NextEntryOffset = binary.LittleEndian.Uint32(data[0:4])
	fi.FileIndex = binary.LittleEndian.Uint32(data[4:8])
	fi.CreationTime = binary.LittleEndian.Uint64(data[8:16])
	fi.LastAccessTime = binary.LittleEndian.Uint64(data[16:24])
	fi.LastWriteTime = binary.LittleEndian.Uint64(data[24:32])
	fi.ChangeTime = binary.LittleEndian.Uint64(data[32:40])
	fi.EndOfFile = int64(binary.LittleEndian.Uint64(data[40:48]))
	fi.AllocationSize = int64(binary.LittleEndian.Uint64(data[48:56]))
	fi.FileAttributes = binary.LittleEndian.Uint32(data[56:60])
	nameLen := int(binary.LittleEndian.Uint32(data[60:64]))
	fi.EaSize = binary.LittleEndian.Uint32(data[64:68])
	shortLen := int(data[68])
	if shortLen > 24 {
		shortLen = 24
	}
	fi.ShortName = decodeUTF16LE(data[70 : 70+shortLen])
	fi.FileId = binary.LittleEndian.Uint64(data[96:104])
	if nameLen > 0 && fileIdBothDirInfoFixedSize+nameLen <= len(data) {
		fi.FileName = decodeUTF16LE(data[fileIdBothDirInfoFixedSize : fileIdBothDirInfoFixedSize+nameLen])
	}
	return nil
}

// ParseFileIdBothDirectoryInformationList decodes a chained buffer of
// FILE_ID_BOTH_DIR_INFORMATION entries.
func ParseFileIdBothDirectoryInformationList(buf []byte) []FileIdBothDirectoryInformation {
	var out []FileIdBothDirectoryInformation
	for pos := 0; pos+fileIdBothDirInfoFixedSize <= len(buf); {
		var e FileIdBothDirectoryInformation
		if err := e.Unmarshal(buf[pos:]); err != nil {
			break
		}
		out = append(out, e)
		if e.NextEntryOffset == 0 {
			break
		}
		pos += int(e.NextEntryOffset)
	}
	return out
}

// FileNamesInformation is FILE_NAMES_INFORMATION (FileInformationClass 12): a
// directory entry carrying only the file name.
//
// [MS-FSCC] 2.4.26 FileNamesInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/a289f7a8-83d2-4927-8c88-b2d328dde5b5
type FileNamesInformation struct {
	NextEntryOffset uint32
	FileIndex       uint32
	FileName        string
}

// fileNamesInfoFixedSize is the fixed portion preceding FileName.
const fileNamesInfoFixedSize = 12

// Marshal encodes a single entry (fixed prefix + FileName).
func (fi *FileNamesInformation) Marshal() ([]byte, error) {
	name := encodeUTF16LE(fi.FileName)
	b := make([]byte, fileNamesInfoFixedSize+len(name))
	binary.LittleEndian.PutUint32(b[0:4], fi.NextEntryOffset)
	binary.LittleEndian.PutUint32(b[4:8], fi.FileIndex)
	binary.LittleEndian.PutUint32(b[8:12], uint32(len(name)))
	copy(b[fileNamesInfoFixedSize:], name)
	return b, nil
}

// Unmarshal decodes a single entry from the start of data.
func (fi *FileNamesInformation) Unmarshal(data []byte) error {
	if len(data) < fileNamesInfoFixedSize {
		return fmt.Errorf("FILE_NAMES_INFORMATION: need %d bytes, got %d", fileNamesInfoFixedSize, len(data))
	}
	fi.NextEntryOffset = binary.LittleEndian.Uint32(data[0:4])
	fi.FileIndex = binary.LittleEndian.Uint32(data[4:8])
	nameLen := int(binary.LittleEndian.Uint32(data[8:12]))
	if nameLen > 0 && fileNamesInfoFixedSize+nameLen <= len(data) {
		fi.FileName = decodeUTF16LE(data[fileNamesInfoFixedSize : fileNamesInfoFixedSize+nameLen])
	}
	return nil
}

// ParseFileNamesInformationList decodes a chained buffer of
// FILE_NAMES_INFORMATION entries.
func ParseFileNamesInformationList(buf []byte) []FileNamesInformation {
	var out []FileNamesInformation
	for pos := 0; pos+fileNamesInfoFixedSize <= len(buf); {
		var e FileNamesInformation
		if err := e.Unmarshal(buf[pos:]); err != nil {
			break
		}
		out = append(out, e)
		if e.NextEntryOffset == 0 {
			break
		}
		pos += int(e.NextEntryOffset)
	}
	return out
}
