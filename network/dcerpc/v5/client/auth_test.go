package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func TestSetAuthValidation(t *testing.T) {
	creds, err := credentials.NewCredentials("", "user", "pass", "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}

	cases := []struct {
		name     string
		authType uint8
		level    uint8
		creds    *credentials.Credentials
		wantErr  bool
	}{
		{"ntlm privacy", pdu.AuthTypeNTLMSSP, pdu.AuthLevelPktPrivacy, creds, false},
		{"ntlm integrity", pdu.AuthTypeNTLMSSP, pdu.AuthLevelPktIntegrity, creds, false},
		{"ntlm pkt", pdu.AuthTypeNTLMSSP, pdu.AuthLevelPkt, creds, false},
		{"ntlm call (promoted to pkt)", pdu.AuthTypeNTLMSSP, pdu.AuthLevelCall, creds, false},
		{"ntlm connect", pdu.AuthTypeNTLMSSP, pdu.AuthLevelConnect, creds, false},
		{"unsupported auth type", pdu.AuthTypeGSSKerberos, pdu.AuthLevelPktPrivacy, creds, true},
		{"unsupported level (none)", pdu.AuthTypeNTLMSSP, pdu.AuthLevelNone, creds, true},
		{"unsupported level (default)", pdu.AuthTypeNTLMSSP, pdu.AuthLevelDefault, creds, true},
		{"nil credentials", pdu.AuthTypeNTLMSSP, pdu.AuthLevelPktPrivacy, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(nil)
			err := c.SetAuth(tc.authType, tc.level, tc.creds)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if c.authConfigured() {
					t.Fatal("auth must not be configured after a rejected SetAuth")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !c.authConfigured() {
				t.Fatal("auth should be configured after a successful SetAuth")
			}
		})
	}
}

// TestSetAuthCallPromotedToPkt verifies that the connection-oriented client maps
// RPC_C_AUTHN_LEVEL_CALL onto PKT, which is what is carried in the sec_trailer.
func TestSetAuthCallPromotedToPkt(t *testing.T) {
	creds, err := credentials.NewCredentials("", "user", "pass", "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}
	c := NewClient(nil)
	if err := c.SetAuth(pdu.AuthTypeNTLMSSP, pdu.AuthLevelCall, creds); err != nil {
		t.Fatalf("SetAuth(call): %v", err)
	}
	if c.authLevel != pdu.AuthLevelPkt {
		t.Errorf("auth level = %d, want PKT (%d) after CALL promotion", c.authLevel, pdu.AuthLevelPkt)
	}
}
