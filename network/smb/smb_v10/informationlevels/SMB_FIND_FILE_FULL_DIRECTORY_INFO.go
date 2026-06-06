package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SMB_FIND_FILE_FULL_DIRECTORY_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/64bd690e-d3a4-4588-96ef-0cb90b065d08
type SMB_FIND_FILE_FULL_DIRECTORY_INFO struct {
	// NextEntryOffset: (4 bytes): This field contains the offset, in bytes, from this
	// entry in the list to the next entry in the list. If there are no additional
	// entries, the value MUST be zero (0x00000000).
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
	// LastAttrChangeTime: (8 bytes): This field contains the date and time when the
	// file attributes where last changed.
	Lastattrchangetime types.FILETIME
	// EndOfFile: (8 bytes): This field contains the offset, in bytes, from the start
	// of the file to the first byte after the end of the file.
	Endoffile types.LARGE_INTEGER
	// AllocationSize: (8 bytes): This field contains the file allocation size, in
	// bytes. Usually, this value is a multiple of the sector or cluster size of the
	// underlying physical device.
	Allocationsize types.LARGE_INTEGER
	// ExtFileAttributes: (4 bytes): This field contains the extended file attributes
	// of the file, encoded as an SMB_EXT_FILE_ATTR (section 2.2.1.2.3) data type.
	Extfileattributes types.SMB_EXT_FILE_ATTR
	// FileNameLength: (4 bytes): This field contains the length of the FileName field,
	// in bytes.
	Filenamelength types.ULONG
	// EaSize: (4 bytes): This field contains the size of the file's extended attribute
	// (EA) information, in bytes.
	Easize types.ULONG
	// FileName: (variable): This field contains the name of the file. It is
	// FileNameLength bytes long; the raw bytes are stored as-is.
	Filename []types.UCHAR
}

// Marshal serializes the SMB_FIND_FILE_FULL_DIRECTORY_INFO into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_FIND_FILE_FULL_DIRECTORY_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// NextEntryOffset (4 bytes) and FileIndex (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Nextentryoffset))
	marshalled_struct = append(marshalled_struct, buf4...)
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Fileindex))
	marshalled_struct = append(marshalled_struct, buf4...)

	// CreationTime, LastAccessTime, LastWriteTime, LastAttrChangeTime (8 bytes each, FILETIME).
	for _, ft := range []*types.FILETIME{&s.Creationtime, &s.Lastaccesstime, &s.Lastwritetime, &s.Lastattrchangetime} {
		b, err := ft.Marshal()
		if err != nil {
			return nil, err
		}
		marshalled_struct = append(marshalled_struct, b...)
	}

	// EndOfFile (8 bytes) and AllocationSize (8 bytes), LARGE_INTEGER.
	buf8 := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf8, uint64(s.Endoffile.QuadPart))
	marshalled_struct = append(marshalled_struct, buf8...)
	buf8 = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf8, uint64(s.Allocationsize.QuadPart))
	marshalled_struct = append(marshalled_struct, buf8...)

	// ExtFileAttributes (4 bytes).
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Extfileattributes))
	marshalled_struct = append(marshalled_struct, buf4...)

	// FileNameLength (4 bytes), derived from the FileName payload.
	s.Filenamelength = types.ULONG(len(s.Filename))
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Filenamelength))
	marshalled_struct = append(marshalled_struct, buf4...)

	// EaSize (4 bytes).
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Easize))
	marshalled_struct = append(marshalled_struct, buf4...)

	// FileName (variable).
	marshalled_struct = append(marshalled_struct, s.Filename...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_FIND_FILE_FULL_DIRECTORY_INFO structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_FIND_FILE_FULL_DIRECTORY_INFO structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_FIND_FILE_FULL_DIRECTORY_INFO) Unmarshal(data []byte) (int, error) {
	// Fixed portion: NextEntryOffset(4) + FileIndex(4) + 4x FILETIME(32) + EndOfFile(8)
	// + AllocationSize(8) + ExtFileAttributes(4) + FileNameLength(4) + EaSize(4) = 68 bytes.
	if len(data) < 68 {
		return 0, fmt.Errorf("data too short for SMB_FIND_FILE_FULL_DIRECTORY_INFO fixed fields (need 68 bytes, have %d)", len(data))
	}
	s.Nextentryoffset = types.ULONG(binary.LittleEndian.Uint32(data[0:4]))
	s.Fileindex = types.ULONG(binary.LittleEndian.Uint32(data[4:8]))
	offset := 8
	for _, ft := range []*types.FILETIME{&s.Creationtime, &s.Lastaccesstime, &s.Lastwritetime, &s.Lastattrchangetime} {
		n, err := ft.Unmarshal(data[offset:])
		if err != nil {
			return offset, err
		}
		offset += n
	}
	s.Endoffile.QuadPart = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	s.Allocationsize.QuadPart = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	s.Extfileattributes = types.SMB_EXT_FILE_ATTR(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	s.Filenamelength = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	s.Easize = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	// FileName (FileNameLength bytes).
	if len(data) < offset+int(s.Filenamelength) {
		return offset, fmt.Errorf("data too short for FileName (need %d bytes, have %d)", s.Filenamelength, len(data)-offset)
	}
	s.Filename = append([]types.UCHAR{}, data[offset:offset+int(s.Filenamelength)]...)
	offset += int(s.Filenamelength)

	return offset, nil
}
