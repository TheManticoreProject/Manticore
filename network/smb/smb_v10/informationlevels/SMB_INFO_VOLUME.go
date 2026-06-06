package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SMB_INFO_VOLUME
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/13d589f5-67e9-49e8-8c33-7b04b8f7cd8c
type SMB_INFO_VOLUME struct {
	// ulVolSerialNbr: (4 bytes): This field contains the serial number of the volume.
	Ulvolserialnbr types.ULONG
	// cCharCount: (1 byte): This field contains the number of characters in the
	// VolumeLabel field.
	Ccharcount types.UCHAR
	// VolumeLabel: (variable): This field contains the volume label. On the wire it is
	// cCharCount bytes with no buffer-format prefix, so it is held as a raw byte slice
	// rather than a buffer-format-prefixed SMB_STRING.
	Volumelabel []types.UCHAR
}

// Marshal serializes the SMB_INFO_VOLUME into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_INFO_VOLUME) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// ulVolSerialNbr (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Ulvolserialnbr))
	marshalled_struct = append(marshalled_struct, buf4...)

	// cCharCount (1 byte), derived from the VolumeLabel payload.
	s.Ccharcount = types.UCHAR(len(s.Volumelabel))
	marshalled_struct = append(marshalled_struct, byte(s.Ccharcount))

	// VolumeLabel (cCharCount bytes).
	marshalled_struct = append(marshalled_struct, s.Volumelabel...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_INFO_VOLUME structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_INFO_VOLUME structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_INFO_VOLUME) Unmarshal(data []byte) (int, error) {
	// ulVolSerialNbr(4) + cCharCount(1) = 5 bytes.
	if len(data) < 5 {
		return 0, fmt.Errorf("data too short for SMB_INFO_VOLUME fixed fields (need 5 bytes, have %d)", len(data))
	}
	s.Ulvolserialnbr = types.ULONG(binary.LittleEndian.Uint32(data[0:4]))
	s.Ccharcount = types.UCHAR(data[4])
	offset := 5

	// VolumeLabel (cCharCount bytes).
	if len(data) < offset+int(s.Ccharcount) {
		return offset, fmt.Errorf("data too short for VolumeLabel (need %d bytes, have %d)", s.Ccharcount, len(data)-offset)
	}
	s.Volumelabel = append([]types.UCHAR{}, data[offset:offset+int(s.Ccharcount)]...)
	offset += int(s.Ccharcount)

	return offset, nil
}
