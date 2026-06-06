package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SMB_INFO_QUERY_EA_SIZE
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/b0a5faf7-e7cc-4b38-878d-2253a2ef9ad4
type SMB_INFO_QUERY_EA_SIZE struct {
	// CreationDate: (2 bytes): This field contains the date when the file was created.
	Creationdate types.SMB_DATE
	// CreationTime: (2 bytes): This field contains the time when the file was created.
	Creationtime types.SMB_TIME_DOS
	// LastAccessDate: (2 bytes): This field contains the date when the file was last
	// accessed.
	Lastaccessdate types.SMB_DATE
	// LastAccessTime: (2 bytes): This field contains the time when the file was last
	// accessed.
	Lastaccesstime types.SMB_TIME_DOS
	// LastWriteDate : (2 bytes): This field contains the date when data was last
	// written to the file.
	Lastwritedate types.SMB_DATE
	// LastWriteTime: (2 bytes): This field contains the time when data was last
	// written to the file.
	Lastwritetime types.SMB_TIME_DOS
	// FileDataSize: (4 bytes): This field contains the file size, in filesystem
	// allocation units.
	Filedatasize types.ULONG
	// AllocationSize: (4 bytes): This field contains the size of the filesystem
	// allocation unit, in bytes.
	Allocationsize types.ULONG
	// Attributes: (2 bytes): This field contains the file attributes.
	Attributes types.SMB_FILE_ATTRIBUTES
	// EaSize: (4 bytes): This field contains the size of the file's extended attribute
	// (EA) information in bytes.
	Easize types.ULONG
}

// Marshal serializes the SMB_INFO_QUERY_EA_SIZE into a byte slice (26 bytes).
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_INFO_QUERY_EA_SIZE) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// Six 2-byte DOS date/time fields, in order.
	dateTimeFields := []infoStandardField{
		&s.Creationdate, &s.Creationtime,
		&s.Lastaccessdate, &s.Lastaccesstime,
		&s.Lastwritedate, &s.Lastwritetime,
	}
	for _, f := range dateTimeFields {
		b, err := f.Marshal()
		if err != nil {
			return nil, err
		}
		marshalled_struct = append(marshalled_struct, b...)
	}

	// FileDataSize (4 bytes) and AllocationSize (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Filedatasize))
	marshalled_struct = append(marshalled_struct, buf4...)
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Allocationsize))
	marshalled_struct = append(marshalled_struct, buf4...)

	// Attributes (2 bytes, SMB_FILE_ATTRIBUTES).
	attrBytes, err := s.Attributes.Marshal()
	if err != nil {
		return nil, err
	}
	marshalled_struct = append(marshalled_struct, attrBytes...)

	// EaSize (4 bytes).
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Easize))
	marshalled_struct = append(marshalled_struct, buf4...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_INFO_QUERY_EA_SIZE structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_INFO_QUERY_EA_SIZE structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_INFO_QUERY_EA_SIZE) Unmarshal(data []byte) (int, error) {
	// 6x 2-byte date/time (12) + FileDataSize(4) + AllocationSize(4) + Attributes(2) + EaSize(4) = 26 bytes.
	if len(data) < 26 {
		return 0, fmt.Errorf("data too short for SMB_INFO_QUERY_EA_SIZE (need 26 bytes, have %d)", len(data))
	}

	dateTimeFields := []infoStandardField{
		&s.Creationdate, &s.Creationtime,
		&s.Lastaccessdate, &s.Lastaccesstime,
		&s.Lastwritedate, &s.Lastwritetime,
	}
	offset := 0
	for _, f := range dateTimeFields {
		n, err := f.Unmarshal(data[offset:])
		if err != nil {
			return offset, err
		}
		offset += n
	}

	s.Filedatasize = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	s.Allocationsize = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if _, err := s.Attributes.Unmarshal(data[offset : offset+2]); err != nil {
		return offset, err
	}
	offset += 2
	s.Easize = types.ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	return offset, nil
}
