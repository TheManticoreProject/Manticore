package securityfeatures

// SecurityFeatures is an interface that represents the 8-byte security features field in the SMB header
// This interface allows for different implementations of the security features field based on the context:
// - SecurityFeaturesSecuritySignature: Used when SMB signing has been negotiated
// - SecurityFeaturesConnectionlessTransport: Used for connectionless transports
// - SecurityFeaturesReserved: Used when neither of the above applies
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/69a29f73-de0c-45a6-a1aa-8ceeea42217f
type SecurityFeatures interface {
	// Marshal serializes the security features into a byte slice
	Marshal() ([]byte, error)

	// Unmarshal deserializes a byte slice into the security features
	Unmarshal([]byte) (int, error)
}
