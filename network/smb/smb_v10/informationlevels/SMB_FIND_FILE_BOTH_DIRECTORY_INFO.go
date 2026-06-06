package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// shortNameSize is the fixed size, in bytes, of the ShortName field of an
// SMB_FIND_FILE_BOTH_DIRECTORY_INFO entry (12 WCHARs).
const shortNameSize = 24

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
	// ShortName: (24 bytes): This field MUST contain the 8.3 name, if any, of the file
	// in Unicode format. It is a fixed 24-byte field; ShortNameLength gives the number
	// of meaningful bytes.
	Shortname [shortNameSize]types.UCHAR
	// FileName: (variable): This field contains the long name of the file. It is
	// FileNameLength bytes long; the raw bytes are stored as-is.
	Filename []types.UCHAR
}

// Marshal serializes the SMB_FIND_FILE_BOTH_DIRECTORY_INFO into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_FIND_FILE_BOTH_DIRECTORY_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// NextEntryOffset (4 bytes) and FileIndex (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Nextentryoffset))
	marshalled_struct = append(marshalled_struct, buf4...)
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Fileindex))
	marshalled_struct = append(marshalled_struct, buf4...)

	// CreationTime, LastAccessTime, LastWriteTime, LastChangeTime (8 bytes each, FILETIME).
	for _, ft := range []*types.FILETIME{&s.Creationtime, &s.Lastaccesstime, &s.Lastwritetime, &s.Lastchangetime} {
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

	// ShortNameLength (1 byte) and Reserved (1 byte).
	marshalled_struct = append(marshalled_struct, byte(s.Shortnamelength), byte(s.Reserved))

	// ShortName (fixed 24 bytes).
	for i := 0; i < shortNameSize; i++ {
		marshalled_struct = append(marshalled_struct, byte(s.Shortname[i]))
	}

	// FileName (variable).
	marshalled_struct = append(marshalled_struct, s.Filename...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_FIND_FILE_BOTH_DIRECTORY_INFO structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_FIND_FILE_BOTH_DIRECTORY_INFO structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_FIND_FILE_BOTH_DIRECTORY_INFO) Unmarshal(data []byte) (int, error) {
	// Fixed portion: NextEntryOffset(4) + FileIndex(4) + 4x FILETIME(32) + EndOfFile(8)
	// + AllocationSize(8) + ExtFileAttributes(4) + FileNameLength(4) + EaSize(4)
	// + ShortNameLength(1) + Reserved(1) + ShortName(24) = 94 bytes.
	const fixedSize = 8 + 32 + 16 + 4 + 4 + 4 + 1 + 1 + shortNameSize // 94
	if len(data) < fixedSize {
		return 0, fmt.Errorf("data too short for SMB_FIND_FILE_BOTH_DIRECTORY_INFO fixed fields (need %d bytes, have %d)", fixedSize, len(data))
	}
	s.Nextentryoffset = types.ULONG(binary.LittleEndian.Uint32(data[0:4]))
	s.Fileindex = types.ULONG(binary.LittleEndian.Uint32(data[4:8]))
	offset := 8
	for _, ft := range []*types.FILETIME{&s.Creationtime, &s.Lastaccesstime, &s.Lastwritetime, &s.Lastchangetime} {
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
	s.Shortnamelength = types.UCHAR(data[offset])
	offset++
	s.Reserved = types.UCHAR(data[offset])
	offset++
	for i := 0; i < shortNameSize; i++ {
		s.Shortname[i] = types.UCHAR(data[offset+i])
	}
	offset += shortNameSize

	// FileName (FileNameLength bytes).
	if len(data) < offset+int(s.Filenamelength) {
		return offset, fmt.Errorf("data too short for FileName (need %d bytes, have %d)", s.Filenamelength, len(data)-offset)
	}
	s.Filename = append([]types.UCHAR{}, data[offset:offset+int(s.Filenamelength)]...)
	offset += int(s.Filenamelength)

	return offset, nil
}
