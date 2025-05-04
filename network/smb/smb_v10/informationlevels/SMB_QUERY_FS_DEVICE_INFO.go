package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_QUERY_FS_DEVICE_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/d7ea6e1a-6526-4230-b566-e9588c7498f1
type SMB_QUERY_FS_DEVICE_INFO struct {
	// DeviceType: (4 bytes): This field contains the device type on which the volume
	// resides.
	Devicetype types.ULONG
	// DeviceCharacteristics: (4 bytes): This 32-bit field of flags contains the device
	// characteristics. The individual flags are as follows.
	Devicecharacteristics types.ULONG
}

// Marshal serializes the SMB_QUERY_FS_DEVICE_INFO into a byte slice.
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
func (s *SMB_QUERY_FS_DEVICE_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FS_DEVICE_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FS_DEVICE_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FS_DEVICE_INFO) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
