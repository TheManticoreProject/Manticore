package informationlevels

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SMB_QUERY_FS_VOLUME_INFO
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/879f3ae2-b029-4b3b-8043-c830fc517b28
type SMB_QUERY_FS_VOLUME_INFO struct {
	// VolumeCreationTime: (8 bytes): This field contains the date and time when the
	// volume was created.
	Volumecreationtime types.FILETIME
	// SerialNumber: (4 bytes): This field contains the serial number of the volume.
	Serialnumber types.ULONG
	// VolumeLabelSize: (4 bytes): This field contains the size of the VolumeLabel
	// field, in bytes.
	Volumelabelsize types.ULONG
	// Reserved: (2 bytes): Reserved.
	Reserved types.USHORT
	// VolumeLabel: (variable): This field contains the Unicode-encoded (UTF-16LE)
	// volume label. It is VolumeLabelSize bytes long; the raw bytes are stored as-is.
	Volumelabel []types.UCHAR
}

// Marshal serializes the SMB_QUERY_FS_VOLUME_INFO into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled information level structure
// - An error if marshalling any component fails
func (s *SMB_QUERY_FS_VOLUME_INFO) Marshal() ([]byte, error) {
	marshalled_struct := []byte{}

	// VolumeCreationTime (8 bytes, FILETIME).
	b, err := s.Volumecreationtime.Marshal()
	if err != nil {
		return nil, err
	}
	marshalled_struct = append(marshalled_struct, b...)

	// SerialNumber (4 bytes).
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Serialnumber))
	marshalled_struct = append(marshalled_struct, buf4...)

	// VolumeLabelSize (4 bytes), derived from the VolumeLabel payload.
	s.Volumelabelsize = types.ULONG(len(s.Volumelabel))
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(s.Volumelabelsize))
	marshalled_struct = append(marshalled_struct, buf4...)

	// Reserved (2 bytes).
	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(s.Reserved))
	marshalled_struct = append(marshalled_struct, buf2...)

	// VolumeLabel (variable).
	marshalled_struct = append(marshalled_struct, s.Volumelabel...)

	return marshalled_struct, nil
}

// Unmarshal deserializes a byte slice into the SMB_QUERY_FS_VOLUME_INFO structure.
//
// Parameters:
// - data: A byte slice containing the serialized SMB_QUERY_FS_VOLUME_INFO structure
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling any component fails or if the data format is invalid
func (s *SMB_QUERY_FS_VOLUME_INFO) Unmarshal(data []byte) (int, error) {
	// VolumeCreationTime(8) + SerialNumber(4) + VolumeLabelSize(4) + Reserved(2) = 18 bytes.
	if len(data) < 18 {
		return 0, fmt.Errorf("data too short for SMB_QUERY_FS_VOLUME_INFO fixed fields (need 18 bytes, have %d)", len(data))
	}
	if _, err := s.Volumecreationtime.Unmarshal(data[0:8]); err != nil {
		return 0, err
	}
	s.Serialnumber = types.ULONG(binary.LittleEndian.Uint32(data[8:12]))
	s.Volumelabelsize = types.ULONG(binary.LittleEndian.Uint32(data[12:16]))
	s.Reserved = types.USHORT(binary.LittleEndian.Uint16(data[16:18]))
	offset := 18

	// VolumeLabel (VolumeLabelSize bytes).
	if len(data) < offset+int(s.Volumelabelsize) {
		return offset, fmt.Errorf("data too short for VolumeLabel (need %d bytes, have %d)", s.Volumelabelsize, len(data)-offset)
	}
	s.Volumelabel = append([]types.UCHAR{}, data[offset:offset+int(s.Volumelabelsize)]...)
	offset += int(s.Volumelabelsize)

	return offset, nil
}
