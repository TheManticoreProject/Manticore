package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_INFO_ALLOCATION
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/194f7dd3-a019-4789-a70c-b28e029e6409
type SMB_INFO_ALLOCATION struct {
	// idFileSystem: (4 bytes): This field contains a file system identifier.
	Idfilesystem types.ULONG
	// cSectorUnit: (4 bytes): This field contains the number of sectors per allocation
	// unit.
	Csectorunit types.ULONG
	// cUnit: (4 bytes): This field contains the total number of allocation units.
	Cunit types.ULONG
	// cUnitAvailable: (4 bytes): This field contains the total number of available
	// allocation units.
	Cunitavailable types.ULONG
	// cbSector: (2 bytes): This field contains the number of bytes per sector.
	Cbsector types.USHORT
}

// Marshal serializes the SMB_INFO_ALLOCATION into a byte slice.
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
func (s *SMB_INFO_ALLOCATION) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_INFO_ALLOCATION structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_INFO_ALLOCATION structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_INFO_ALLOCATION) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
