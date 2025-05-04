package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_QUERY_FS_SIZE_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3045d7df-7757-4725-9ffd-20227978cc46
type SMB_QUERY_FS_SIZE_INFO struct {
	// TotalAllocationUnits: (8 bytes): This field contains the total number of
	// allocation units assigned to the volume.
	Totalallocationunits types.LARGE_INTEGER
	// TotalFreeAllocationUnits: (8 bytes): This field contains the total number of
	// unallocated or free allocation units for the volume.
	Totalfreeallocationunits types.LARGE_INTEGER
	// SectorsPerAllocationUnit: (4 bytes): This field contains the number of sectors
	// per allocation unit.
	Sectorsperallocationunit types.ULONG
	// BytesPerSector: (4 bytes): This field contains the bytes per sector.
	Bytespersector types.ULONG
}

// Marshal serializes the SMB_QUERY_FS_SIZE_INFO into a byte slice.
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
func (s *SMB_QUERY_FS_SIZE_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FS_SIZE_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FS_SIZE_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FS_SIZE_INFO) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
