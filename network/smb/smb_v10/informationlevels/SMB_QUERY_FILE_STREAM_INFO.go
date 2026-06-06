package informationlevels

import (
	"encoding/binary"
	"fmt"

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
	// StreamName: (variable): A sequence of Unicode (UTF-16LE) characters containing the
	// name of the stream. It is StreamNameLength bytes long and might not be
	// null-terminated; the raw bytes are stored as-is.
	Streamname []types.UCHAR
}

// Marshal serializes the SMB_QUERY_FILE_STREAM_INFO into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_QUERY_FILE_STREAM_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// NextEntryOffset (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Nextentryoffset))
	marshalled_struct = append(marshalled_struct, buf4...)

	// StreamNameLength (4 bytes), derived from the StreamName payload.
	s.Streamnamelength = types.ULONG(len(s.Streamname))
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Streamnamelength))
	marshalled_struct = append(marshalled_struct, buf4...)

	// StreamSize (8 bytes) and StreamAllocationSize (8 bytes), LARGE_INTEGER.
	buf8 := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf8, uint64(s.Streamsize.QuadPart))
	marshalled_struct = append(marshalled_struct, buf8...)
	buf8 = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf8, uint64(s.Streamallocationsize.QuadPart))
	marshalled_struct = append(marshalled_struct, buf8...)

	// StreamName (variable).
	marshalled_struct = append(marshalled_struct, s.Streamname...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FILE_STREAM_INFO structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FILE_STREAM_INFO structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FILE_STREAM_INFO) Unmarshal(data []byte) (int, error) {
	// NextEntryOffset(4) + StreamNameLength(4) + StreamSize(8) + StreamAllocationSize(8) = 24 bytes.
	if len(data) < 24 {
		return 0, fmt.Errorf("data too short for SMB_QUERY_FILE_STREAM_INFO fixed fields (need 24 bytes, have %d)", len(data))
	}
	s.Nextentryoffset = types.ULONG(binary.LittleEndian.Uint32(data[0:4]))
	s.Streamnamelength = types.ULONG(binary.LittleEndian.Uint32(data[4:8]))
	s.Streamsize.QuadPart = binary.LittleEndian.Uint64(data[8:16])
	s.Streamallocationsize.QuadPart = binary.LittleEndian.Uint64(data[16:24])
	offset := 24

	// StreamName (StreamNameLength bytes).
	if len(data) < offset+int(s.Streamnamelength) {
		return offset, fmt.Errorf("data too short for StreamName (need %d bytes, have %d)", s.Streamnamelength, len(data)-offset)
	}
	s.Streamname = append([]types.UCHAR{}, data[offset:offset+int(s.Streamnamelength)]...)
	offset += int(s.Streamnamelength)

	return offset, nil
}
