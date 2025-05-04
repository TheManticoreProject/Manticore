package securityfeatures

import "fmt"

// SecurityFeaturesSecuritySignature represents the 8-byte security features field in the SMB header
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/69a29f73-de0c-45a6-a1aa-8ceeea42217f
type SecurityFeaturesSecuritySignature struct {
	// SecuritySignature (8 bytes): If SMB signing has been negotiated, this field MUST contain an
	// 8-byte cryptographic message signature that can be used to detect whether the message was modified
	// while in transit. The use of message signing is mutually exclusive with connectionless transport.
	SecuritySignature [8]byte
}

// NewSecurityFeaturesSecuritySignature creates a new SecurityFeaturesSecuritySignature with a zeroed 8-byte security signature
func NewSecurityFeaturesSecuritySignature() *SecurityFeaturesSecuritySignature {
	return &SecurityFeaturesSecuritySignature{
		SecuritySignature: [8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
}

// Marshal marshals the 8-byte security signature into a byte slice
func (s *SecurityFeaturesSecuritySignature) Marshal() ([]byte, error) {
	return s.SecuritySignature[:], nil
}

// Unmarshal unmarshals the 8-byte security signature from the given data
func (s *SecurityFeaturesSecuritySignature) Unmarshal(data []byte) (int, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("expected 8 bytes, got %d", len(data))
	}

	copy(s.SecuritySignature[:], data[:8])

	return 8, nil
}

// GetSecuritySignature returns the 8-byte security signature
func (s *SecurityFeaturesSecuritySignature) GetSecuritySignature() [8]byte {
	return s.SecuritySignature
}

// SetSecuritySignature sets the 8-byte security signature
func (s *SecurityFeaturesSecuritySignature) SetSecuritySignature(signature [8]byte) {
	s.SecuritySignature = signature
}
