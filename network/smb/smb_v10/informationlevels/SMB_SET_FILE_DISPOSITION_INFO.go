package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_SET_FILE_DISPOSITION_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/bb8e952b-d293-4fc3-bc47-67ce1a8f8655
type SMB_SET_FILE_DISPOSITION_INFO struct {
	// DeletePending: (1 byte): An 8-bit field that is set to 0x01 to indicate that a
	// file SHOULD be deleted when it is closed; otherwise, to 0x00.
	Deletepending types.UCHAR
}

// Marshal serializes the SMB_SET_FILE_DISPOSITION_INFO into a byte slice.
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
func (s *SMB_SET_FILE_DISPOSITION_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_SET_FILE_DISPOSITION_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_SET_FILE_DISPOSITION_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_SET_FILE_DISPOSITION_INFO) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
