package types

import (
	"encoding/binary"
	"fmt"
)

// SMB_FEA_LIST
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/1ca1684e-6552-432c-bdd0-f559814bbaef
type SMB_FEA_LIST struct {
	// SizeOfListInBytes (4 bytes): The total size of the FEAList field plus the size
	// of the SizeOfListInBytes field itself (4 bytes).
	SizeOfListInBytes ULONG
	// FEAList (variable): A concatenated list of SMB_FEA structures, held here as raw
	// bytes.
	FEAList []UCHAR
}

// Marshal serializes the SMB_FEA_LIST into a byte slice.
//
// Returns:
// - A byte slice containing the marshalled SMB_FEA_LIST structure
// - An error if marshalling fails
func (s *SMB_FEA_LIST) Marshal() ([]byte, error) {
	// SizeOfListInBytes counts the 4-byte size field plus the FEAList payload.
	s.SizeOfListInBytes = ULONG(4 + len(s.FEAList))

	marshalled := make([]byte, 4)
	binary.LittleEndian.PutUint32(marshalled, uint32(s.SizeOfListInBytes))
	marshalled = append(marshalled, s.FEAList...)

	return marshalled, nil
}

// Unmarshal deserializes a byte slice into the SMB_FEA_LIST structure.
//
// Parameters:
// - data: The byte slice to unmarshal
//
// Returns:
// - The number of bytes consumed
// - An error if unmarshalling fails or the data format is invalid
func (s *SMB_FEA_LIST) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short for SMB_FEA_LIST SizeOfListInBytes")
	}
	s.SizeOfListInBytes = ULONG(binary.LittleEndian.Uint32(data[0:4]))

	// SizeOfListInBytes includes the 4-byte size field itself.
	if s.SizeOfListInBytes < 4 {
		return 0, fmt.Errorf("SMB_FEA_LIST SizeOfListInBytes (%d) is smaller than the 4-byte header", s.SizeOfListInBytes)
	}
	payloadLen := int(s.SizeOfListInBytes) - 4
	if len(data) < 4+payloadLen {
		return 0, fmt.Errorf("data too short for SMB_FEA_LIST payload (need %d bytes, have %d)", payloadLen, len(data)-4)
	}
	s.FEAList = append([]UCHAR{}, data[4:4+payloadLen]...)

	return 4 + payloadLen, nil
}
