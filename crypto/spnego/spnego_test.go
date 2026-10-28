package spnego_test

import (
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate"
	negotiate_flags "github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/types"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
)

func TestCreateNegTokenInit(t *testing.T) {
	// Create a simple NTLM NEGOTIATE message
	negotiateFlags := negotiate_flags.NegotiateFlags(
		negotiate_flags.NTLMSSP_NEGOTIATE_NTLM |
			negotiate_flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
			negotiate_flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
			negotiate_flags.NTLMSSP_NEGOTIATE_128 |
			negotiate_flags.NTLMSSP_NEGOTIATE_56 |
			negotiate_flags.NTLMSSP_REQUEST_TARGET |
			negotiate_flags.NTLMSSP_NEGOTIATE_TARGET_INFO,
	)

	ntlmNegotiate, err := negotiate.CreateNegotiateMessage("DOMAIN", "WORKSTATION", negotiateFlags, nil)
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
	negotiateFlags := negotiate_flags.NegotiateFlags(
		negotiate_flags.NTLMSSP_NEGOTIATE_NTLM |
			negotiate_flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
			negotiate_flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
			negotiate_flags.NTLMSSP_NEGOTIATE_128 |
			negotiate_flags.NTLMSSP_NEGOTIATE_56 |
			negotiate_flags.NTLMSSP_REQUEST_TARGET |
			negotiate_flags.NTLMSSP_NEGOTIATE_TARGET_INFO |
			negotiate_flags.NTLMSSP_NEGOTIATE_VERSION,
	)
	v := &version.Version{
		ProductMajorVersion: 5,
		ProductMinorVersion: 1,
		ProductBuild:        0,
		Reserved:            [3]byte{0, 0, 0},
		NTLMRevision:        version.NTLMSSP_REVISION_W2K3,
	}
	token, err := ctx.CreateNegotiateToken(negotiateFlags, v)
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
	if types.MessageType(messageType) != types.MESSAGE_TYPE_NEGOTIATE {
		t.Errorf("Expected message type 1 (NEGOTIATE), got %d", messageType)
	}
}

func TestStripGSSHeaderRejectsOverflowingLength(t *testing.T) {
	for _, token := range [][]byte{
		{0x60, 0x88, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		{0x60, 0x84, 0xFF, 0xFF, 0xFF, 0xFF, 0x00},
		{0x60, 0x84, 0x7F, 0xFF, 0xFF, 0xFF, 0x00},
		{0x60, 0x85, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
	} {
		if _, err := spnego.ParseNegTokenInit(token); err == nil {
			t.Errorf("ParseNegTokenInit(% x) accepted an overflowing length", token)
		}
		if _, err := spnego.ExtractNTLMToken(token); err == nil {
			t.Errorf("ExtractNTLMToken(% x) accepted an overflowing length", token)
		}
	}
}
