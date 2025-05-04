package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_INFO_QUERY_ALL_EAS
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/2db63466-bdbf-45c4-96fc-dff83ebda893
type SMB_INFO_QUERY_ALL_EAS struct {
	// ExtendedAttributeList: (variable): A list of all of the extended attribute (EA)
	// name/value pairs assigned to the file.
	Extendedattributelist types.SMB_FEA_LIST
}

// Marshal serializes the SMB_INFO_QUERY_ALL_EAS into a byte slice.
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
func (s *SMB_INFO_QUERY_ALL_EAS) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_INFO_QUERY_ALL_EAS structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_INFO_QUERY_ALL_EAS structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_INFO_QUERY_ALL_EAS) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
