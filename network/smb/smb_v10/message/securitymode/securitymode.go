package securitymode

type SecurityMode uint8

const (
	SecurityModeNone SecurityMode = iota
)

// Security mode flags for SMB negotiation
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/a4229e1a-8a4e-489a-a2eb-11b7f360e60c
const (
	// NEGOTIATE_USER_SECURITY: If clear (0), the server supports only Share Level access control.
	// If set (1), the server supports only User Level access control.
	NEGOTIATE_USER_SECURITY SecurityMode = 0x01

	// NEGOTIATE_ENCRYPT_PASSWORDS: If clear, the server supports only plaintext password authentication.
	// If set, the server supports challenge/response authentication.
	NEGOTIATE_ENCRYPT_PASSWORDS SecurityMode = 0x02

	// NEGOTIATE_SECURITY_SIGNATURES_ENABLED: If clear, the server does not support SMB security signatures.
	// If set, the server supports SMB security signatures for this connection.
	NEGOTIATE_SECURITY_SIGNATURES_ENABLED SecurityMode = 0x04

	// NEGOTIATE_SECURITY_SIGNATURES_REQUIRED: If clear, the security signatures are optional for this connection.
	// If set, the server requires security signatures.
	// This bit MUST be clear if the NEGOTIATE_SECURITY_SIGNATURES_ENABLED bit is clear.
	NEGOTIATE_SECURITY_SIGNATURES_REQUIRED SecurityMode = 0x08

	// Reserved: The remaining bits (0xF0) are reserved and MUST be zero.
)

// SupportsPlaintextPasswordAuth returns true if the server supports plaintext password authentication
func (sm SecurityMode) SupportsPlaintextPasswordAuth() bool {
	return (sm & NEGOTIATE_ENCRYPT_PASSWORDS) == 0
}

// SupportsChallengeResponseAuth returns true if the server supports challenge/response authentication
func (sm SecurityMode) SupportsChallengeResponseAuth() bool {
	return (sm & NEGOTIATE_ENCRYPT_PASSWORDS) != 0
}

// SupportsShareLevelAccessControl returns true if the server supports Share Level access control
func (sm SecurityMode) SupportsShareLevelAccessControl() bool {
	return (sm & NEGOTIATE_USER_SECURITY) == 0
}

// SupportsUserLevelAccessControl returns true if the server supports User Level access control
func (sm SecurityMode) SupportsUserLevelAccessControl() bool {
	return (sm & NEGOTIATE_USER_SECURITY) != 0
}

// IsSecuritySignatureEnabled returns true if the server supports SMB security signatures
func (sm SecurityMode) IsSecuritySignatureEnabled() bool {
	return (sm & NEGOTIATE_SECURITY_SIGNATURES_ENABLED) != 0
}

// IsSecuritySignatureRequired returns true if the server requires security signatures
func (sm SecurityMode) IsSecuritySignatureRequired() bool {
	return (sm & NEGOTIATE_SECURITY_SIGNATURES_REQUIRED) != 0
}
