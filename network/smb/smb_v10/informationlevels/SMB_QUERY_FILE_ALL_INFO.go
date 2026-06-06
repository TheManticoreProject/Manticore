package informationlevels

import (
	"encoding/binary"
	"fmt"

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
	// AllocationSize: (8 bytes): This field contains the number of bytes that are
	// allocated to the file.
	Allocationsize types.LARGE_INTEGER
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
	// FileName: (variable): This field contains the name of the file in Unicode
	// (UTF-16LE). It is FileNameLength bytes long; the raw bytes are stored as-is.
	Filename []types.UCHAR
}

// Marshal serializes the SMB_QUERY_FILE_ALL_INFO into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_QUERY_FILE_ALL_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// CreationTime, LastAccessTime, LastWriteTime, LastChangeTime (8 bytes each, FILETIME).
	for _, ft := range []*types.FILETIME{&s.Creationtime, &s.Lastaccesstime, &s.Lastwritetime, &s.Lastchangetime} {
		b, err := ft.Marshal()
		if err != nil {
			return nil, err
		}
		marshalled_struct = append(marshalled_struct, b...)
	}

	// ExtFileAttributes (4 bytes) and Reserved1 (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Extfileattributes))
	marshalled_struct = append(marshalled_struct, buf4...)
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Reserved1))
	marshalled_struct = append(marshalled_struct, buf4...)

	// AllocationSize (8 bytes) and EndOfFile (8 bytes), LARGE_INTEGER.
	buf8 := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf8, uint64(s.Allocationsize.QuadPart))
	marshalled_struct = append(marshalled_struct, buf8...)
	buf8 = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf8, uint64(s.Endoffile.QuadPart))
	marshalled_struct = append(marshalled_struct, buf8...)

	// NumberOfLinks (4 bytes).
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Numberoflinks))
	marshalled_struct = append(marshalled_struct, buf4...)

	// DeletePending (1 byte) and Directory (1 byte).
	marshalled_struct = append(marshalled_struct, byte(s.Deletepending), byte(s.Directory))

	// Reserved2 (2 bytes).
	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(s.Reserved2))
	marshalled_struct = append(marshalled_struct, buf2...)

	// EaSize (4 bytes).
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Easize))
	marshalled_struct = append(marshalled_struct, buf4...)

	// FileNameLength (4 bytes), derived from the FileName payload.
	s.Filenamelength = types.ULONG(len(s.Filename))
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Filenamelength))
	marshalled_struct = append(marshalled_struct, buf4...)

	// FileName (variable).
	marshalled_struct = append(marshalled_struct, s.Filename...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FILE_ALL_INFO structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FILE_ALL_INFO structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FILE_ALL_INFO) Unmarshal(data []byte) (int, error) {
	offset := 0

	// CreationTime, LastAccessTime, LastWriteTime, LastChangeTime (8 bytes each, FILETIME).
	for _, ft := range []*types.FILETIME{&s.Creationtime, &s.Lastaccesstime, &s.Lastwritetime, &s.Lastchangetime} {
		n, err := ft.Unmarshal(data[offset:])
		if err != nil {
			return offset, err
		}
		offset += n
	}

	// Fixed fields up to and including FileNameLength: ExtFileAttributes(4) + Reserved1(4)
	// + AllocationSize(8) + EndOfFile(8) + NumberOfLinks(4) + DeletePending(1) + Directory(1)
	// + Reserved2(2) + EaSize(4) + FileNameLength(4) = 40 bytes after the four FILETIMEs.
	if len(data) < offset+40 {
		return offset, fmt.Errorf("data too short for SMB_QUERY_FILE_ALL_INFO fixed fields")
	}
	s.Extfileattributes = types.SMB_EXT_FILE_ATTR(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	s.Reserved1 = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	s.Allocationsize.QuadPart = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	s.Endoffile.QuadPart = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	s.Numberoflinks = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	s.Deletepending = types.UCHAR(data[offset])
	offset++
	s.Directory = types.UCHAR(data[offset])
	offset++
	s.Reserved2 = types.USHORT(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	s.Easize = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	s.Filenamelength = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	// FileName (FileNameLength bytes).
	if len(data) < offset+int(s.Filenamelength) {
		return offset, fmt.Errorf("data too short for FileName (need %d bytes, have %d)", s.Filenamelength, len(data)-offset)
	}
	s.Filename = append([]types.UCHAR{}, data[offset:offset+int(s.Filenamelength)]...)
	offset += int(s.Filenamelength)

	return offset, nil
}
