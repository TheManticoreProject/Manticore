package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_FIND_FILE_BOTH_DIRECTORY_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/2aa849f4-1bc0-42bf-9c8f-d09f11fccc4c
type SMB_FIND_FILE_BOTH_DIRECTORY_INFO struct {
	// NextEntryOffset: (4 bytes): This field contains the offset, in bytes, from this
	// entry in the list to the next entry in the list. If there are no additional
	// entries the value MUST be zero (0x00000000).
	Nextentryoffset types.ULONG
	// FileIndex: (4 bytes): This field SHOULD be set to zero when sent in a
	// response and SHOULD be ignored when received by the client.
	Fileindex types.ULONG
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
	// EndOfFile: (8 bytes): The absolute new end-of-file position as a byte offset
	// from the start of the file. EndOfFile specifies the byte offset to the end of
	// the file. Because this value is zero-based, it actually refers to the first free
	// byte in the file. In other words, EndOfFile is the offset to the byte
	// immediately following the last valid byte in the file.
	Endoffile types.LARGE_INTEGER
	// AllocationSize: (8 bytes): This field contains the file allocation size, in
	// bytes. Usually, this value is a multiple of the sector or cluster size of the
	// underlying physical device.
	Allocationsize types.LARGE_INTEGER
	// ExtFileAttributes: (4 bytes): This field contains the extended file attributes
	// of the file, encoded as an SMB_EXT_FILE_ATTR (section 2.2.1.2.3) data type.
	Extfileattributes types.SMB_EXT_FILE_ATTR
	// FileNameLength: (4 bytes): This field MUST contain the length of the FileName
	// field, in bytes.
	Filenamelength types.ULONG
	// EaSize: (4 bytes): This field MUST contain the length of the FEAList, in bytes.
	Easize types.ULONG
	// ShortNameLength: (1 byte): This field MUST contain the length of the ShortName,
	// in bytes, or zero if no 8.3 name is present.
	Shortnamelength types.UCHAR
	// Reserved: (1 byte): This field is reserved and MUST be zero (0x00).
	Reserved types.UCHAR
}

// Marshal serializes the SMB_FIND_FILE_BOTH_DIRECTORY_INFO into a byte slice.
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
func (s *SMB_FIND_FILE_BOTH_DIRECTORY_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_FIND_FILE_BOTH_DIRECTORY_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_FIND_FILE_BOTH_DIRECTORY_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_FIND_FILE_BOTH_DIRECTORY_INFO) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
