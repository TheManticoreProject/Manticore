package securityfeatures

import (
	"encoding/binary"
	"fmt"
)

// SecurityFeatures represents the 8-byte security features field in the SMB header
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/69a29f73-de0c-45a6-a1aa-8ceeea42217f
type SecurityFeaturesConnectionlessTransport struct {
	// Key (4 bytes): An encryption key used for validating messages over connectionless transports.
	Key uint32
	// CID (2 bytes): A connection identifier (CID).
	CID uint16
	// SequenceNumber (2 bytes): A number used to identify the sequence of a message over connectionless transports.
	SequenceNumber uint16
}

func NewSecurityFeaturesConnectionlessTransport() *SecurityFeaturesConnectionlessTransport {
	return &SecurityFeaturesConnectionlessTransport{
		Key:            0x00000000,
		CID:            0x0000,
		SequenceNumber: 0x0000,
	}
}

func (s *SecurityFeaturesConnectionlessTransport) Marshal() ([]byte, error) {
	buf := []byte{}

	bufUint32 := make([]byte, 4)
	binary.LittleEndian.PutUint32(bufUint32, s.Key)
	buf = append(buf, bufUint32...)

	bufUint16 := make([]byte, 2)
	binary.LittleEndian.PutUint16(bufUint16, s.CID)
	buf = append(buf, bufUint16...)

	bufUint16 = make([]byte, 2)
	binary.LittleEndian.PutUint16(bufUint16, s.SequenceNumber)
	buf = append(buf, bufUint16...)

	return buf, nil
}

func (s *SecurityFeaturesConnectionlessTransport) Unmarshal(data []byte) (int, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("invalid security features length: %d", len(data))
	}

	s.Key = binary.LittleEndian.Uint32(data[:4])

	s.CID = binary.LittleEndian.Uint16(data[4:6])

	s.SequenceNumber = binary.LittleEndian.Uint16(data[6:8])

	return 8, nil
}
