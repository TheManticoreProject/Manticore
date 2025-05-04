package securityfeatures

import "fmt"

// SecurityFeaturesReserved represents the 8-byte security features field in the SMB header
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/69a29f73-de0c-45a6-a1aa-8ceeea42217f
type SecurityFeaturesReserved struct {
	// Finally, if neither of the above two cases applies,the SecurityFeatures field is treated as a reserved field,
	// which MUST be set to zero by the client and MUST be ignored by the server.
	Reserved [8]byte
}

func NewSecurityFeaturesReserved() *SecurityFeaturesReserved {
	return &SecurityFeaturesReserved{
		Reserved: [8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
}

func (s *SecurityFeaturesReserved) Marshal() ([]byte, error) {
	return s.Reserved[:], nil
}

func (s *SecurityFeaturesReserved) Unmarshal(data []byte) (int, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("invalid security features length: %d", len(data))
	}

	copy(s.Reserved[:], data[:8])

	return 8, nil
}
