package informationlevels

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SMB_INFO_STANDARD
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3e6f3a13-6a40-4f76-af70-bb514554ea5b
type SMB_INFO_STANDARD struct {
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
	// LastWriteDate: (2 bytes): This field contains the date when data was last
	// written to the file.
	Lastwritedate types.SMB_DATE
	// LastWriteTime: (2 bytes): This field contains the time when data was last
	// written to the file.
	Lastwritetime types.SMB_TIME_DOS
}

// marshaler is the common 2-byte (de)serialization interface implemented by
// SMB_DATE and SMB_TIME_DOS.
type infoStandardField interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) (int, error)
}

// Marshal serializes the SMB_INFO_STANDARD into a byte slice (12 bytes).
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_INFO_STANDARD) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// Six 2-byte DOS date/time fields, in order.
	fields := []infoStandardField{
		&s.Creationdate, &s.Creationtime,
		&s.Lastaccessdate, &s.Lastaccesstime,
		&s.Lastwritedate, &s.Lastwritetime,
	}
	for _, f := range fields {
		b, err := f.Marshal()
		if err != nil {
			return nil, err
		}
		marshalled_struct = append(marshalled_struct, b...)
	}

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_INFO_STANDARD structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_INFO_STANDARD structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_INFO_STANDARD) Unmarshal(data []byte) (int, error) {
	// Six 2-byte DOS date/time fields = 12 bytes.
	if len(data) < 12 {
		return 0, fmt.Errorf("data too short for SMB_INFO_STANDARD (need 12 bytes, have %d)", len(data))
	}

	fields := []infoStandardField{
		&s.Creationdate, &s.Creationtime,
		&s.Lastaccessdate, &s.Lastaccesstime,
		&s.Lastwritedate, &s.Lastwritetime,
	}
	offset := 0
	for _, f := range fields {
		n, err := f.Unmarshal(data[offset:])
		if err != nil {
			return offset, err
		}
		offset += n
	}

	return offset, nil
}
