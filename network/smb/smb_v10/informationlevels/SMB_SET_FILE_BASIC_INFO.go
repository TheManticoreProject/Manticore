package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)


// SMB_SET_FILE_BASIC_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/021549da-ef78-4282-ae93-7ae93acaba97
type SMB_SET_FILE_BASIC_INFO struct {
	// CreationTime: (8 bytes): A 64-bit unsigned integer that contains the time when
	// the file was created. A valid time for this field is an integer greater than
	// 0x0000000000000000. When setting file attributes, a value of 0x0000000000000000
	// indicates to the server that it MUST NOT change this attribute. When setting
	// file attributes, a value of -1 (0xFFFFFFFFFFFFFFFF) indicates to the server that
	// it MUST NOT change this attribute for all subsequent operations on the same file
	// handle. This field MUST NOT be set to a value less than -1 (0xFFFFFFFFFFFFFFFF).
	Creationtime types.FILETIME
	// LastAccessTime: (8 bytes): A 64-bit unsigned integer that contains the last time
	// that the file was accessed, in the format of a FILETIME structure. A valid time
	// for this field is an integer greater than 0x0000000000000000. When setting file
	// attributes, a value of 0x0000000000000000 indicates to the server that it MUST
	// NOT change this attribute. When setting file attributes, a value of -1
	// (0xFFFFFFFFFFFFFFFF) indicates to the server that it MUST NOT change this
	// attribute for all subsequent operations on the same file handle. This field MUST
	// NOT be set to a value less than -1 (0xFFFFFFFFFFFFFFFF).
	Lastaccesstime types.FILETIME
	// LastWriteTime: (8 bytes): A 64-bit unsigned integer that contains the last time
	// that information was written to the file, in the format of a FILETIME structure.
	// A valid time for this field is an integer greater than 0x0000000000000000. When
	// setting file attributes, a value of 0x0000000000000000 indicates to the server
	// that it MUST NOT change this attribute. When setting file attributes, a value of
	// -1 (0xFFFFFFFFFFFFFFFF) indicates to the server that it MUST NOT change this
	// attribute for all subsequent operations on the same file handle. This field MUST
	// NOT be set to a value less than -1 (0xFFFFFFFFFFFFFFFF).
	Lastwritetime types.FILETIME
	// ChangeTime: (8 bytes): A 64-bit unsigned integer that contains the last time
	// that the file was changed, in the format of a FILETIME structure. A valid time
	// for this field is an integer greater than 0x0000000000000000. When setting file
	// attributes, a value of 0x0000000000000000 indicates to the server that it MUST
	// NOT change this attribute. When setting file attributes, a value of -1
	// (0xFFFFFFFFFFFFFFFF) indicates to the server that it MUST NOT change this
	// attribute for all subsequent operations on the same file handle. This field MUST
	// NOT be set to a value less than -1 (0xFFFFFFFFFFFFFFFF).
	Changetime types.FILETIME
	// ExtFileAttributes: (4 bytes): This field contains the extended file attributes
	// of the file, encoded as an SMB_EXT_FILE_ATTR data type (section 2.2.1.2.3).
	Extfileattributes types.SMB_EXT_FILE_ATTR
	// Reserved: (4 bytes): A 32-bit reserved field that can be set to any value and
	// MUST be ignored.
	Reserved types.ULONG
}

// Marshal serializes the SMB_SET_FILE_BASIC_INFO into a byte slice.
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
func (s *SMB_SET_FILE_BASIC_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// CreationTime, LastAccessTime, LastWriteTime, ChangeTime (8 bytes each, FILETIME).
	for _, ft := range []*types.FILETIME{&s.Creationtime, &s.Lastaccesstime, &s.Lastwritetime, &s.Changetime} {
		b, err := ft.Marshal()
		if err != nil {
			return nil, err
		}
		marshalled_struct = append(marshalled_struct, b...)
	}

	// ExtFileAttributes (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Extfileattributes))
	marshalled_struct = append(marshalled_struct, buf4...)

	// Reserved (4 bytes).
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Reserved))
	marshalled_struct = append(marshalled_struct, buf4...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_SET_FILE_BASIC_INFO structure.
//
// This method unmarshals the information level structure according to the format
// specified in MS-CIFS documentation. Information levels are used in various
// SMB operations to determine the format of data being exchanged.
//
// The data is expected to follow the specific format required for this information level.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_SET_FILE_BASIC_INFO structure
//
// Returns:
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_SET_FILE_BASIC_INFO) Unmarshal(data []byte) (int, error) {
	offset := 0

	// CreationTime, LastAccessTime, LastWriteTime, ChangeTime (8 bytes each, FILETIME).
	for _, ft := range []*types.FILETIME{&s.Creationtime, &s.Lastaccesstime, &s.Lastwritetime, &s.Changetime} {
		n, err := ft.Unmarshal(data[offset:])
		if err != nil {
			return offset, err
		}
		offset += n
	}

	// ExtFileAttributes (4 bytes) and Reserved (4 bytes).
	if len(data) < offset+8 {
		return offset, fmt.Errorf("data too short for ExtFileAttributes and Reserved")
	}
	s.Extfileattributes = types.SMB_EXT_FILE_ATTR(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	s.Reserved = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	return offset, nil
}
