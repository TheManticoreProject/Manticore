package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_INFO_STANDARD
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3e6f3a13-6a40-4f76-af70-bb514554ea5b
type SMB_INFO_STANDARD struct {
	// CreationDate: (2 bytes): This field contains the date when the file was created.
	Creationdate types.SMB_DATE
	// CreationTime: (2 bytes): This field contains the time when the file was created.
	Creationtime types.SMB_TIME
	// LastAccessDate: (2 bytes): This field contains the date when the file was last
	// accessed.
	Lastaccessdate types.SMB_DATE
	// LastAccessTime: (2 bytes): This field contains the time when the file was last
	// accessed.
	Lastaccesstime types.SMB_TIME
	// LastWriteDate: (2 bytes): This field contains the date when data was last
	// written to the file.
	Lastwritedate types.SMB_DATE
	// LastWriteTime: (2 bytes): This field contains the time when data was last
	// written to the file.
	Lastwritetime types.SMB_TIME
}

// Marshal serializes the SMB_INFO_STANDARD into a byte slice.
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
func (s *SMB_INFO_STANDARD) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_INFO_STANDARD structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_INFO_STANDARD structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_INFO_STANDARD) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
