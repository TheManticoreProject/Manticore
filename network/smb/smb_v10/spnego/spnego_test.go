package spnego_test

import (
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/negotiate"
)

func TestCreateNegTokenInit(t *testing.T) {
	// Create a simple NTLM NEGOTIATE message
	ntlmNegotiate, err := negotiate.CreateNegotiateMessage("DOMAIN", "WORKSTATION", true)
	if err != nil {
		t.Fatalf("Failed to create NTLM NEGOTIATE message: %v", err)
	}

	ntlmNegotiateBytes, err := ntlmNegotiate.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal NTLM NEGOTIATE message: %v", err)
	}

	// Wrap it in SPNEGO
	token, err := spnego.CreateNegTokenInit(ntlmNegotiateBytes)
	if err != nil {
		t.Fatalf("Failed to create SPNEGO token: %v", err)
	}

	// Verify the token starts with the GSS-API header
	if token[0] != spnego.GSS_API_SPNEGO {
		t.Errorf("Expected token to start with GSS-API header 0x60, got 0x%02x", token[0])
	}

	// Extract the NTLM token
	extractedToken, err := spnego.ExtractNTLMToken(token)
	if err != nil {
		t.Fatalf("Failed to extract NTLM token: %v", err)
	}

	// Verify it's the same as the original
	if len(extractedToken) != len(ntlmNegotiateBytes) {
		t.Errorf("Extracted token length %d doesn't match original %d", len(extractedToken), len(ntlmNegotiateBytes))
	}

	// Check the NTLM signature
	if string(extractedToken[:8]) != "NTLMSSP\x00" {
		t.Errorf("Expected NTLM signature, got %s", hex.EncodeToString(extractedToken[:8]))
	}
}

func TestAuthContext(t *testing.T) {
	// Create an auth context
	ctx := spnego.NewAuthContext(
		spnego.AuthTypeNTLM,
		"DOMAIN",
		"user",
		"password",
		"WORKSTATION",
		true,
	)

	// Create negotiate token
	token, err := ctx.CreateNegotiateToken()
	if err != nil {
		t.Fatalf("Failed to create negotiate token: %v", err)
	}

	// Verify the token
	if token[0] != spnego.GSS_API_SPNEGO {
		t.Errorf("Expected token to start with GSS-API header 0x60, got 0x%02x", token[0])
	}

	// Extract and verify the NTLM token
	ntlmToken, err := spnego.ExtractNTLMToken(token)
	if err != nil {
		t.Fatalf("Failed to extract NTLM token: %v", err)
	}

	if string(ntlmToken[:8]) != "NTLMSSP\x00" {
		t.Errorf("Expected NTLM signature, got %s", hex.EncodeToString(ntlmToken[:8]))
	}

	// Verify message type is NEGOTIATE (1)
	messageType := ntlmToken[8]
	if messageType != 1 {
		t.Errorf("Expected message type 1 (NEGOTIATE), got %d", messageType)
	}
}
