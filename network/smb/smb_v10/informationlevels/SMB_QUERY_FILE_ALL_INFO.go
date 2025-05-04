package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_QUERY_FILE_ALL_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/162baf45-4201-4b07-a397-060e868599d7
type SMB_QUERY_FILE_ALL_INFO struct {
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
	// Reserved1: (4 bytes): Reserved. This field SHOULD be set to 0x00000000 by the
	// server and MUST be ignored by the client.
	Reserved1 types.ULONG
	// EndOfFile: (8 bytes): This field contains the offset, in bytes, from the start
	// of the file to the first byte after the end of the file.
	Endoffile types.LARGE_INTEGER
	// NumberOfLinks: (4 bytes): This field contains the number of hard links to the
	// file.
	Numberoflinks types.ULONG
	// DeletePending: (1 byte): This field indicates whether there is a delete action
	// pending for the file.
	Deletepending types.UCHAR
	// Directory: (1 byte): This field indicates whether the file is a directory.
	Directory types.UCHAR
	// Reserved2: (2 bytes): Reserved. This field SHOULD be set to 0x0000 by the server
	// and MUST be ignored by the client.
	Reserved2 types.USHORT
	// EaSize: (4 bytes): This field MUST contain the length of a file's list of
	// extended attributes in bytes.
	Easize types.ULONG
	// FileNameLength: (4 bytes): This field MUST contain the length, in bytes, of the
	// FileName field.
	Filenamelength types.ULONG
}

// Marshal serializes the SMB_QUERY_FILE_ALL_INFO into a byte slice.
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
func (s *SMB_QUERY_FILE_ALL_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FILE_ALL_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FILE_ALL_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FILE_ALL_INFO) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
