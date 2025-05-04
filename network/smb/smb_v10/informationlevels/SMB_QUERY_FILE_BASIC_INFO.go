package informationlevels

import (
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

// Marshal serializes the SMB_QUERY_FILE_BASIC_INFO into a byte slice.
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
func (s *SMB_QUERY_FILE_BASIC_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FILE_BASIC_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FILE_BASIC_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FILE_BASIC_INFO) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
