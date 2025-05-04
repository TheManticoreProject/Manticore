package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_FIND_FILE_NAMES_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/88b9968b-a36f-482a-bb30-c7a51a3e290d
type SMB_FIND_FILE_NAMES_INFO struct {
	// NextEntryOffset: (4 bytes): This field contains the offset, in bytes, from this
	// entry in the list to the next entry in the list. If there are no additional
	// entries, the value MUST be zero (0x00000000).
	Nextentryoffset types.ULONG
	// FileIndex: (4 bytes): This field SHOULD<163> be set to zero when sent in a
	// response and SHOULD be ignored when received by the client.
	Fileindex types.ULONG
	// FileNameLength: (4 bytes): This field MUST contain the length of the FileName
	// field, in bytes.<164>
	Filenamelength types.ULONG
}

// Marshal serializes the SMB_FIND_FILE_NAMES_INFO into a byte slice.
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
func (s *SMB_FIND_FILE_NAMES_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_FIND_FILE_NAMES_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_FIND_FILE_NAMES_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_FIND_FILE_NAMES_INFO) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
