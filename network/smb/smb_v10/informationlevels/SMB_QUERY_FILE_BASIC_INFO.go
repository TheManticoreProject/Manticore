package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SMB_QUERY_FILE_BASIC_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3da7df75-43ba-4498-a6b3-a68ba57ec922
type SMB_QUERY_FILE_BASIC_INFO struct {
	// CreationTime: (8 bytes): This field contains the date and time when the file was
	// created.
	Creationtime types.FILETIME
	// LastAccessTime: (8 bytes): This field contains the date and time when the file
	// was last accessed.
	Lastaccesstime types.FILETIME
	// LastWriteTime: (8 bytes): This field contains the date and time when data was
	// last written to the file.
	Lastwritetime types.FILETIME
	// LastChangeTime: (8 bytes): This field contains the date and time when the file
	// was last changed.
	Lastchangetime types.FILETIME
	// ExtFileAttributes: (4 bytes): This field contains the extended file attributes
	// of the file, encoded as an SMB_EXT_FILE_ATTR (section 2.2.1.2.3) data type.
	Extfileattributes types.SMB_EXT_FILE_ATTR
	// Reserved: (4 bytes): MUST be set to zero when sent and MUST be ignored on
	// receipt.
	Reserved types.ULONG
}

// Marshal serializes the SMB_QUERY_FILE_BASIC_INFO into a byte slice (40 bytes).
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_QUERY_FILE_BASIC_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// CreationTime, LastAccessTime, LastWriteTime, LastChangeTime (8 bytes each, FILETIME).
	for _, ft := range []*types.FILETIME{&s.Creationtime, &s.Lastaccesstime, &s.Lastwritetime, &s.Lastchangetime} {
		b, err := ft.Marshal()
		if err != nil {
			return nil, err
		}
		marshalled_struct = append(marshalled_struct, b...)
	}

	// ExtFileAttributes (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Extfileattributes))
	marshalled_struct = append(marshalled_struct, buf4...)

	// Reserved (4 bytes).
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Reserved))
	marshalled_struct = append(marshalled_struct, buf4...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FILE_BASIC_INFO structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FILE_BASIC_INFO structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FILE_BASIC_INFO) Unmarshal(data []byte) (int, error) {
	offset := 0

	// CreationTime, LastAccessTime, LastWriteTime, LastChangeTime (8 bytes each, FILETIME).
	for _, ft := range []*types.FILETIME{&s.Creationtime, &s.Lastaccesstime, &s.Lastwritetime, &s.Lastchangetime} {
		n, err := ft.Unmarshal(data[offset:])
		if err != nil {
			return offset, err
		}
		offset += n
	}

	// ExtFileAttributes (4 bytes) and Reserved (4 bytes).
	if len(data) < offset+8 {
		return offset, fmt.Errorf("data too short for ExtFileAttributes and Reserved")
	}
	s.Extfileattributes = types.SMB_EXT_FILE_ATTR(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	s.Reserved = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	return offset, nil
}
