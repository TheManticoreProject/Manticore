package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SMB_QUERY_FILE_ALT_NAME_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3edd12e7-f407-4b46-9465-c6ed20e24c1a
type SMB_QUERY_FILE_ALT_NAME_INFO struct {
	// FileNameLength: (4 bytes): This field contains the length, in bytes, of the
	// FileName field.
	Filenamelength types.ULONG
	// FileName: (variable): This field contains the 8.3 name of the file in Unicode
	// (UTF-16LE). It is FileNameLength bytes long and is not null-terminated. The raw
	// bytes are stored as-is.
	Filename []types.UCHAR
}

// Marshal serializes the SMB_QUERY_FILE_ALT_NAME_INFO into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_QUERY_FILE_ALT_NAME_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// FileNameLength (4 bytes), derived from the FileName payload.
	s.Filenamelength = types.ULONG(len(s.Filename))
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Filenamelength))
	marshalled_struct = append(marshalled_struct, buf4...)

	// FileName (variable).
	marshalled_struct = append(marshalled_struct, s.Filename...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FILE_ALT_NAME_INFO structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FILE_ALT_NAME_INFO structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FILE_ALT_NAME_INFO) Unmarshal(data []byte) (int, error) {
	// FileNameLength (4 bytes).
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short for FileNameLength")
	}
	s.Filenamelength = types.ULONG(binary.LittleEndian.Uint32(data[0:4]))
	offset := 4

	// FileName (FileNameLength bytes).
	if len(data) < offset+int(s.Filenamelength) {
		return offset, fmt.Errorf("data too short for FileName (need %d bytes, have %d)", s.Filenamelength, len(data)-offset)
	}
	s.Filename = append([]types.UCHAR{}, data[offset:offset+int(s.Filenamelength)]...)
	offset += int(s.Filenamelength)

	return offset, nil
}
