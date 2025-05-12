package securitymode_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/securitymode"
)

func TestSecurityModeFlags(t *testing.T) {
	tests := []struct {
		name                          string
		securityMode                  securitymode.SecurityMode
		wantPlaintextPasswordAuth     bool
		wantChallengeResponseAuth     bool
		wantShareLevelAccessControl   bool
		wantUserLevelAccessControl    bool
		wantSecuritySignatureEnabled  bool
		wantSecuritySignatureRequired bool
	}{
		{
			name:                          "None",
			securityMode:                  securitymode.SecurityModeNone,
			wantPlaintextPasswordAuth:     true,
			wantChallengeResponseAuth:     false,
			wantShareLevelAccessControl:   true,
			wantUserLevelAccessControl:    false,
			wantSecuritySignatureEnabled:  false,
			wantSecuritySignatureRequired: false,
		},
		{
			name:                          "User Level Access Control",
			securityMode:                  securitymode.NEGOTIATE_USER_SECURITY,
			wantPlaintextPasswordAuth:     true,
			wantChallengeResponseAuth:     false,
			wantShareLevelAccessControl:   false,
			wantUserLevelAccessControl:    true,
			wantSecuritySignatureEnabled:  false,
			wantSecuritySignatureRequired: false,
		},
		{
			name:                          "Challenge/Response Auth",
			securityMode:                  securitymode.NEGOTIATE_ENCRYPT_PASSWORDS,
			wantPlaintextPasswordAuth:     false,
			wantChallengeResponseAuth:     true,
			wantShareLevelAccessControl:   true,
			wantUserLevelAccessControl:    false,
			wantSecuritySignatureEnabled:  false,
			wantSecuritySignatureRequired: false,
		},
		{
			name:                          "Security Signatures Enabled",
			securityMode:                  securitymode.NEGOTIATE_SECURITY_SIGNATURES_ENABLED,
			wantPlaintextPasswordAuth:     true,
			wantChallengeResponseAuth:     false,
			wantShareLevelAccessControl:   true,
			wantUserLevelAccessControl:    false,
			wantSecuritySignatureEnabled:  true,
			wantSecuritySignatureRequired: false,
		},
		{
			name:                          "Security Signatures Required",
			securityMode:                  securitymode.NEGOTIATE_SECURITY_SIGNATURES_ENABLED | securitymode.NEGOTIATE_SECURITY_SIGNATURES_REQUIRED,
			wantPlaintextPasswordAuth:     true,
			wantChallengeResponseAuth:     false,
			wantShareLevelAccessControl:   true,
			wantUserLevelAccessControl:    false,
			wantSecuritySignatureEnabled:  true,
			wantSecuritySignatureRequired: true,
		},
		{
			name:                          "All Flags Set",
			securityMode:                  securitymode.NEGOTIATE_USER_SECURITY | securitymode.NEGOTIATE_ENCRYPT_PASSWORDS | securitymode.NEGOTIATE_SECURITY_SIGNATURES_ENABLED | securitymode.NEGOTIATE_SECURITY_SIGNATURES_REQUIRED,
			wantPlaintextPasswordAuth:     false,
			wantChallengeResponseAuth:     true,
			wantShareLevelAccessControl:   false,
			wantUserLevelAccessControl:    true,
			wantSecuritySignatureEnabled:  true,
			wantSecuritySignatureRequired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.securityMode.SupportsPlaintextPasswordAuth(); got != tt.wantPlaintextPasswordAuth {
				t.Errorf("SecurityMode.SupportsPlaintextPasswordAuth() = %v, want %v", got, tt.wantPlaintextPasswordAuth)
			}
			if got := tt.securityMode.SupportsChallengeResponseAuth(); got != tt.wantChallengeResponseAuth {
				t.Errorf("SecurityMode.SupportsChallengeResponseAuth() = %v, want %v", got, tt.wantChallengeResponseAuth)
			}
			if got := tt.securityMode.SupportsShareLevelAccessControl(); got != tt.wantShareLevelAccessControl {
				t.Errorf("SecurityMode.SupportsShareLevelAccessControl() = %v, want %v", got, tt.wantShareLevelAccessControl)
			}
			if got := tt.securityMode.SupportsUserLevelAccessControl(); got != tt.wantUserLevelAccessControl {
				t.Errorf("SecurityMode.SupportsUserLevelAccessControl() = %v, want %v", got, tt.wantUserLevelAccessControl)
			}
			if got := tt.securityMode.IsSecuritySignatureEnabled(); got != tt.wantSecuritySignatureEnabled {
				t.Errorf("SecurityMode.IsSecuritySignatureEnabled() = %v, want %v", got, tt.wantSecuritySignatureEnabled)
			}
			if got := tt.securityMode.IsSecuritySignatureRequired(); got != tt.wantSecuritySignatureRequired {
				t.Errorf("SecurityMode.IsSecuritySignatureRequired() = %v, want %v", got, tt.wantSecuritySignatureRequired)
			}
		})
	}
}
