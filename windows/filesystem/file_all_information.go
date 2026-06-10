package filesystem

import (
	"encoding/binary"
	"fmt"
)

// FileAllInformation is FILE_ALL_INFORMATION (FileInformationClass 18): the
// aggregate query result combining the Basic, Standard, Internal, EA, Access,
// Position, Mode, Alignment, and Name information classes for an open file.
//
// The small single-field classes are flattened into scalar fields here:
// IndexNumber (FILE_INTERNAL_INFORMATION), EaSize (FILE_EA_INFORMATION),
// AccessFlags (FILE_ACCESS_INFORMATION), CurrentByteOffset
// (FILE_POSITION_INFORMATION), Mode (FILE_MODE_INFORMATION), and
// AlignmentRequirement (FILE_ALIGNMENT_INFORMATION). FileName is the
// FILE_NAME_INFORMATION value (UTF-16LE on the wire).
//
// [MS-FSCC] 2.4.2 FileAllInformation:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/95f3056a-ebc1-4f5d-b938-3f68a44677a6
type FileAllInformation struct {
	Basic                FileBasicInformation
	Standard             FileStandardInformation
	IndexNumber          int64  // FILE_INTERNAL_INFORMATION
	EaSize               uint32 // FILE_EA_INFORMATION
	AccessFlags          uint32 // FILE_ACCESS_INFORMATION
	CurrentByteOffset    int64  // FILE_POSITION_INFORMATION
	Mode                 uint32 // FILE_MODE_INFORMATION
	AlignmentRequirement uint32 // FILE_ALIGNMENT_INFORMATION
	FileName             string // FILE_NAME_INFORMATION
}

// fileAllInformationFixedSize is the size of FILE_ALL_INFORMATION up to (but not
// including) the variable FileName: Basic(40) Standard(24) Internal(8) EA(4)
// Access(4) Position(8) Mode(4) Alignment(4) FileNameLength(4) = 100.
const fileAllInformationFixedSize = FileBasicInformationSize + FileStandardInformationSize + 8 + 4 + 4 + 8 + 4 + 4 + 4

// Marshal encodes the structure to its wire form (fixed prefix + FileName).
func (fi *FileAllInformation) Marshal() ([]byte, error) {
	name := encodeUTF16LE(fi.FileName)

	basic, err := fi.Basic.Marshal()
	if err != nil {
		return nil, err
	}
	standard, err := fi.Standard.Marshal()
	if err != nil {
		return nil, err
	}

	b := make([]byte, fileAllInformationFixedSize+len(name))
	off := 0
	copy(b[off:], basic)
	off += FileBasicInformationSize
	copy(b[off:], standard)
	off += FileStandardInformationSize
	binary.LittleEndian.PutUint64(b[off:off+8], uint64(fi.IndexNumber))
	off += 8
	binary.LittleEndian.PutUint32(b[off:off+4], fi.EaSize)
	off += 4
	binary.LittleEndian.PutUint32(b[off:off+4], fi.AccessFlags)
	off += 4
	binary.LittleEndian.PutUint64(b[off:off+8], uint64(fi.CurrentByteOffset))
	off += 8
	binary.LittleEndian.PutUint32(b[off:off+4], fi.Mode)
	off += 4
	binary.LittleEndian.PutUint32(b[off:off+4], fi.AlignmentRequirement)
	off += 4
	binary.LittleEndian.PutUint32(b[off:off+4], uint32(len(name)))
	off += 4
	copy(b[off:], name)
	return b, nil
}

// Unmarshal decodes the structure from its wire form. A buffer carrying only the
// fixed prefix (FileNameLength 0 or a truncated name) yields an empty FileName.
func (fi *FileAllInformation) Unmarshal(data []byte) error {
	if len(data) < fileAllInformationFixedSize {
		return fmt.Errorf("FILE_ALL_INFORMATION: need %d bytes, got %d", fileAllInformationFixedSize, len(data))
	}
	off := 0
	if err := fi.Basic.Unmarshal(data[off : off+FileBasicInformationSize]); err != nil {
		return err
	}
	off += FileBasicInformationSize
	if err := fi.Standard.Unmarshal(data[off : off+FileStandardInformationSize]); err != nil {
		return err
	}
	off += FileStandardInformationSize
	fi.IndexNumber = int64(binary.LittleEndian.Uint64(data[off : off+8]))
	off += 8
	fi.EaSize = binary.LittleEndian.Uint32(data[off : off+4])
	off += 4
	fi.AccessFlags = binary.LittleEndian.Uint32(data[off : off+4])
	off += 4
	fi.CurrentByteOffset = int64(binary.LittleEndian.Uint64(data[off : off+8]))
	off += 8
	fi.Mode = binary.LittleEndian.Uint32(data[off : off+4])
	off += 4
	fi.AlignmentRequirement = binary.LittleEndian.Uint32(data[off : off+4])
	off += 4
	nameLen := int(binary.LittleEndian.Uint32(data[off : off+4]))
	off += 4

	if nameLen > 0 && off+nameLen <= len(data) {
		fi.FileName = decodeUTF16LE(data[off : off+nameLen])
	} else {
		fi.FileName = ""
	}
	return nil
}
