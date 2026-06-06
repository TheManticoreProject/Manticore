package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_QUERY_FILE_STANDARD_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3bdd080c-f8a4-4a09-acf1-0f8bd00152e4
type SMB_QUERY_FILE_STANDARD_INFO struct {
	Allocationsize types.LARGE_INTEGER
	// EndOfFile: (8 bytes): This field contains the offset, in bytes, from the start
	// of the file to the first byte after the end of the file.
	Endoffile types.LARGE_INTEGER
	// NumberOfLinks: (4 bytes): This field contains the number of hard links to the
	// file.
	Numberoflinks types.ULONG
	// DeletePending: (1 byte): This field indicates whether there is a delete action
	// pending for the file.
	Deletepending types.UCHAR
	// Directory: (1 byte): This field indicates whether the file is a directory.
	Directory types.UCHAR
}

// Marshal serializes the SMB_QUERY_FILE_STANDARD_INFO into a byte slice.
//
// This method marshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The marshalled data follows the specific format required for this information level.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_QUERY_FILE_STANDARD_INFO) Marshal() ([]byte, error) {
	marshalled_struct := make([]byte, 0, 22)

	// AllocationSize (8 bytes) and EndOfFile (8 bytes), LARGE_INTEGER.
	buf8 := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf8, uint64(s.Allocationsize.QuadPart))
	marshalled_struct = append(marshalled_struct, buf8...)
	buf8 = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf8, uint64(s.Endoffile.QuadPart))
	marshalled_struct = append(marshalled_struct, buf8...)

	// NumberOfLinks (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Numberoflinks))
	marshalled_struct = append(marshalled_struct, buf4...)

	// DeletePending (1 byte) and Directory (1 byte).
	marshalled_struct = append(marshalled_struct, byte(s.Deletepending), byte(s.Directory))

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FILE_STANDARD_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FILE_STANDARD_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FILE_STANDARD_INFO) Unmarshal(data []byte) (int, error) {
	// AllocationSize(8) + EndOfFile(8) + NumberOfLinks(4) + DeletePending(1) + Directory(1) = 22 bytes.
	if len(data) < 22 {
		return 0, fmt.Errorf("data too short for SMB_QUERY_FILE_STANDARD_INFO (need 22 bytes, have %d)", len(data))
	}
	s.Allocationsize.QuadPart = binary.LittleEndian.Uint64(data[0:8])
	s.Endoffile.QuadPart = binary.LittleEndian.Uint64(data[8:16])
	s.Numberoflinks = types.ULONG(binary.LittleEndian.Uint32(data[16:20]))
	s.Deletepending = types.UCHAR(data[20])
	s.Directory = types.UCHAR(data[21])

	return 22, nil
}
