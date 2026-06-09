package types

import (
	"encoding/binary"
	"fmt"
)

// SMB2_FILEID_SIZE is the fixed wire size, in bytes, of an SMB2_FILEID structure.
const SMB2_FILEID_SIZE = 16

// SMB2_FILEID is the 16-byte identifier used by SMB2 to refer to an open file.
// It is returned by an SMB2 CREATE Response and supplied on subsequent file
// operations (CLOSE, READ, WRITE, FLUSH, IOCTL, ...).
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/2dbe6679-c8fd-4e3b-be4c-3b5dd1b58cf2
type SMB2_FILEID struct {
	// Persistent (8 bytes): A file handle that remains persistent when an open
	// is reconnected after being lost; opaque to the client.
	Persistent UINT64

	// Volatile (8 bytes): A file handle that can change when an open is
	// reconnected; opaque to the client.
	Volatile UINT64
}

// Marshal serializes the SMB2_FILEID into its 16-byte little-endian representation.
//
// Returns:
//   - []byte: The serialized file identifier (always 16 bytes).
//   - error: Always nil; present for interface symmetry with other wire types.
func (f *SMB2_FILEID) Marshal() ([]byte, error) {
	buf := make([]byte, SMB2_FILEID_SIZE)
	binary.LittleEndian.PutUint64(buf[0:8], f.Persistent)
	binary.LittleEndian.PutUint64(buf[8:16], f.Volatile)
	return buf, nil
}

// Unmarshal deserializes a 16-byte SMB2_FILEID from the input byte slice.
//
// Parameters:
//   - data: The byte slice containing the serialized file identifier.
//
// Returns:
//   - int: The number of bytes read (16 on success).
//   - error: An error if the input is shorter than 16 bytes.
func (f *SMB2_FILEID) Unmarshal(data []byte) (int, error) {
	if len(data) < SMB2_FILEID_SIZE {
		return 0, fmt.Errorf("data too short to unmarshal SMB2_FILEID: have %d bytes, need %d", len(data), SMB2_FILEID_SIZE)
	}
	f.Persistent = binary.LittleEndian.Uint64(data[0:8])
	f.Volatile = binary.LittleEndian.Uint64(data[8:16])
	return SMB2_FILEID_SIZE, nil
}
