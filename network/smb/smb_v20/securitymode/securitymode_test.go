package securitymode_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

func TestSecurityMode(t *testing.T) {
	both := securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED | securitymode.SMB2_NEGOTIATE_SIGNING_REQUIRED
	if !both.IsSigningEnabled() || !both.IsSigningRequired() {
		t.Errorf("expected both signing bits set for %s", both)
	}
	if securitymode.SecurityMode(0).String() != "NONE" {
		t.Errorf("SecurityMode(0).String() = %q, want NONE", securitymode.SecurityMode(0).String())
	}
	if got := securitymode.SMB2_NEGOTIATE_SIGNING_REQUIRED.String(); got != "SIGNING_REQUIRED" {
		t.Errorf("String() = %q, want SIGNING_REQUIRED", got)
	}
}
