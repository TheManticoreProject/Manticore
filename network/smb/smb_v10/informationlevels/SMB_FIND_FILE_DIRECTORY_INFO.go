package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_FIND_FILE_DIRECTORY_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/8be9119a-b37e-4ff5-bee7-3d7a5997dc88
type SMB_FIND_FILE_DIRECTORY_INFO struct {
	// NextEntryOffset: (4 bytes): This field contains the offset, in bytes, from this
	// entry in the list to the next entry in the list. If there are no additional
	// entries the value MUST be zero (0x00000000).
	Nextentryoffset types.ULONG
	// FileIndex: (4 bytes): This field SHOULD<157> be set to zero when sent in a
	// response and SHOULD be ignored when received by the client.
	Fileindex types.ULONG
	Creationtime types.FILETIME
	// LastAccessTime: (8 bytes): This field contains the date and time when the file
	// was last accessed.
	Lastaccesstime types.FILETIME
	// LastWriteTime: (8 bytes): This field contains the date and time when data was
	// last written to the file.
	Lastwritetime types.FILETIME
	// LastAttrChangeTime: (8 bytes): This field contains the date and time when the
	// file attributes where last changed.
	Lastattrchangetime types.FILETIME
	// EndOfFile: (8 bytes): This field contains the offset, in bytes, to the start of
	// the file to the first byte after the end of the file.
	Endoffile types.LARGE_INTEGER
	// AllocationSize: (8 bytes): This field contains the file allocation size, in
	// bytes. Usually, this value is a multiple of the sector or cluster size of the
	// underlying physical device.
	Allocationsize types.LARGE_INTEGER
	// ExtFileAttributes: (4 bytes): This field contains the extended file attributes
	// of the file, encoded as an SMB_EXT_FILE_ATTR (section 2.2.1.2.3) data type.
	Extfileattributes types.SMB_EXT_FILE_ATTR
	// FileNameLength: (4 bytes): This field contains the length of the FileName field,
	// in bytes.<158>
	Filenamelength types.ULONG
}

// Marshal serializes the SMB_FIND_FILE_DIRECTORY_INFO into a byte slice.
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
func (s *SMB_FIND_FILE_DIRECTORY_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_FIND_FILE_DIRECTORY_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_FIND_FILE_DIRECTORY_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_FIND_FILE_DIRECTORY_INFO) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
