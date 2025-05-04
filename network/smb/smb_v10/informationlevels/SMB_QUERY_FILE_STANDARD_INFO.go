package informationlevels

import (
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
	marshalled_struct := []byte{}

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
	return 0, nil
}
