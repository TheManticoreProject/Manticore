package securitymode

import "strings"

// SecurityMode specifies whether SMB signing is enabled or required. It appears
// as a 2-byte field in SMB2 NEGOTIATE and as a 1-byte field in SMB2
// SESSION_SETUP; the bit values are identical in both.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/e14db7ff-763a-4263-8b10-0c3944f52fc5
type SecurityMode uint16

const (
	// SMB2_NEGOTIATE_SIGNING_ENABLED indicates that security signatures are
	// enabled.
	SMB2_NEGOTIATE_SIGNING_ENABLED SecurityMode = 0x0001
	// SMB2_NEGOTIATE_SIGNING_REQUIRED indicates that security signatures are
	// required.
	SMB2_NEGOTIATE_SIGNING_REQUIRED SecurityMode = 0x0002
)

// IsSigningEnabled reports whether the signing-enabled bit is set.
func (s SecurityMode) IsSigningEnabled() bool {
	return s&SMB2_NEGOTIATE_SIGNING_ENABLED != 0
}

// IsSigningRequired reports whether the signing-required bit is set.
func (s SecurityMode) IsSigningRequired() bool {
	return s&SMB2_NEGOTIATE_SIGNING_REQUIRED != 0
}

// String returns a pipe-separated list of the set mode names, or "NONE".
func (s SecurityMode) String() string {
	var parts []string
	if s.IsSigningEnabled() {
		parts = append(parts, "SIGNING_ENABLED")
	}
	if s.IsSigningRequired() {
		parts = append(parts, "SIGNING_REQUIRED")
	}
	if len(parts) == 0 {
		return "NONE"
	}
	return strings.Join(parts, "|")
}
