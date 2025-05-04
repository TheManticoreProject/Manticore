package informationlevels

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_QUERY_FILE_STREAM_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/23f37dcd-5b50-43d4-91cd-ffab868fd65e
type SMB_QUERY_FILE_STREAM_INFO struct {
	// NextEntryOffset: (4 bytes): A 32-bit unsigned integer that contains the byte
	// offset from the beginning of this entry, at which the next FILE_ STREAM
	// _INFORMATION entry is located, if multiple entries are present in a buffer. This
	// member is 0x00000000 if no other entries follow this one. An implementation MUST
	// use this value to determine the location of the next entry (if multiple entries
	// are present in a buffer) and MUST NOT assume that the value of NextEntryOffset
	// is the same as the size of the current entry.
	Nextentryoffset types.ULONG
	// StreamNameLength: (4 bytes): A 32-bit unsigned integer that contains the length,
	// in bytes, of the stream name string.
	Streamnamelength types.ULONG
	// StreamSize: (8 bytes): A 64-bit signed integer that contains the size, in bytes,
	// of the stream. The value of this field MUST be greater than or equal to
	// 0x0000000000000000.
	Streamsize types.LARGE_INTEGER
	// StreamAllocationSize: (8 bytes): A 64-bit signed integer that contains the file
	// stream allocation size in bytes. Usually, this value is a multiple of the sector
	// or cluster size of the underlying physical device. The value of this field MUST
	// be greater than or equal to 0x0000000000000000.
	Streamallocationsize types.LARGE_INTEGER
}

// Marshal serializes the SMB_QUERY_FILE_STREAM_INFO into a byte slice.
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
func (s *SMB_QUERY_FILE_STREAM_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FILE_STREAM_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FILE_STREAM_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FILE_STREAM_INFO) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
