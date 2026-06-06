package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SMB_QUERY_FS_ATTRIBUTE_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/1011206a-55c5-4dbf-aff0-119514136940
type SMB_QUERY_FS_ATTRIBUTE_INFO struct {
	// FileSystemAttributes: (4 bytes): A 32-bit field of flags that contains the file
	// system's attributes (e.g. FILE_CASE_SENSITIVE_SEARCH, FILE_UNICODE_ON_DISK).
	Filesystemattributes types.ULONG
	// MaxFileNameLengthInBytes: (4 bytes): The maximum size, in bytes, of a file name
	// on the file system.
	Maxfilenamelengthinbytes types.ULONG
	// LengthOfFileSystemName: (4 bytes): The size, in bytes, of the FileSystemName
	// field.
	Lengthoffilesystemname types.ULONG
	// FileSystemName: (variable): The Unicode-encoded (UTF-16LE) name of the file
	// system. It is LengthOfFileSystemName bytes long; the raw bytes are stored as-is.
	Filesystemname []types.UCHAR
}

// Marshal serializes the SMB_QUERY_FS_ATTRIBUTE_INFO into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_QUERY_FS_ATTRIBUTE_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// FileSystemAttributes (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Filesystemattributes))
	marshalled_struct = append(marshalled_struct, buf4...)

	// MaxFileNameLengthInBytes (4 bytes).
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Maxfilenamelengthinbytes))
	marshalled_struct = append(marshalled_struct, buf4...)

	// LengthOfFileSystemName (4 bytes), derived from the FileSystemName payload.
	s.Lengthoffilesystemname = types.ULONG(len(s.Filesystemname))
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Lengthoffilesystemname))
	marshalled_struct = append(marshalled_struct, buf4...)

	// FileSystemName (variable).
	marshalled_struct = append(marshalled_struct, s.Filesystemname...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FS_ATTRIBUTE_INFO structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FS_ATTRIBUTE_INFO structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FS_ATTRIBUTE_INFO) Unmarshal(data []byte) (int, error) {
	// FileSystemAttributes(4) + MaxFileNameLengthInBytes(4) + LengthOfFileSystemName(4) = 12 bytes.
	if len(data) < 12 {
		return 0, fmt.Errorf("data too short for SMB_QUERY_FS_ATTRIBUTE_INFO fixed fields (need 12 bytes, have %d)", len(data))
	}
	s.Filesystemattributes = types.ULONG(binary.LittleEndian.Uint32(data[0:4]))
	s.Maxfilenamelengthinbytes = types.ULONG(binary.LittleEndian.Uint32(data[4:8]))
	s.Lengthoffilesystemname = types.ULONG(binary.LittleEndian.Uint32(data[8:12]))
	offset := 12

	// FileSystemName (LengthOfFileSystemName bytes).
	if len(data) < offset+int(s.Lengthoffilesystemname) {
		return offset, fmt.Errorf("data too short for FileSystemName (need %d bytes, have %d)", s.Lengthoffilesystemname, len(data)-offset)
	}
	s.Filesystemname = append([]types.UCHAR{}, data[offset:offset+int(s.Lengthoffilesystemname)]...)
	offset += int(s.Lengthoffilesystemname)

	return offset, nil
}
