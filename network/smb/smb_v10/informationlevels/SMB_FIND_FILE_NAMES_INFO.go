package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SMB_FIND_FILE_NAMES_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/88b9968b-a36f-482a-bb30-c7a51a3e290d
type SMB_FIND_FILE_NAMES_INFO struct {
	// NextEntryOffset: (4 bytes): This field contains the offset, in bytes, from this
	// entry in the list to the next entry in the list. If there are no additional
	// entries, the value MUST be zero (0x00000000).
	Nextentryoffset types.ULONG
	// FileIndex: (4 bytes): This field SHOULD be set to zero when sent in a
	// response and SHOULD be ignored when received by the client.
	Fileindex types.ULONG
	// FileNameLength: (4 bytes): This field MUST contain the length of the FileName
	// field, in bytes.
	Filenamelength types.ULONG
	// FileName: (variable): This field contains the name of the file. It is
	// FileNameLength bytes long; the raw bytes are stored as-is.
	Filename []types.UCHAR
}

// Marshal serializes the SMB_FIND_FILE_NAMES_INFO into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_FIND_FILE_NAMES_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// NextEntryOffset (4 bytes) and FileIndex (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Nextentryoffset))
	marshalled_struct = append(marshalled_struct, buf4...)
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Fileindex))
	marshalled_struct = append(marshalled_struct, buf4...)

	// FileNameLength (4 bytes), derived from the FileName payload.
	s.Filenamelength = types.ULONG(len(s.Filename))
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Filenamelength))
	marshalled_struct = append(marshalled_struct, buf4...)

	// FileName (variable).
	marshalled_struct = append(marshalled_struct, s.Filename...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_FIND_FILE_NAMES_INFO structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_FIND_FILE_NAMES_INFO structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_FIND_FILE_NAMES_INFO) Unmarshal(data []byte) (int, error) {
	// NextEntryOffset(4) + FileIndex(4) + FileNameLength(4) = 12 bytes.
	if len(data) < 12 {
		return 0, fmt.Errorf("data too short for SMB_FIND_FILE_NAMES_INFO fixed fields")
	}
	s.Nextentryoffset = types.ULONG(binary.LittleEndian.Uint32(data[0:4]))
	s.Fileindex = types.ULONG(binary.LittleEndian.Uint32(data[4:8]))
	s.Filenamelength = types.ULONG(binary.LittleEndian.Uint32(data[8:12]))
	offset := 12

	// FileName (FileNameLength bytes).
	if len(data) < offset+int(s.Filenamelength) {
		return offset, fmt.Errorf("data too short for FileName (need %d bytes, have %d)", s.Filenamelength, len(data)-offset)
	}
	s.Filename = append([]types.UCHAR{}, data[offset:offset+int(s.Filenamelength)]...)
	offset += int(s.Filenamelength)

	return offset, nil
}
