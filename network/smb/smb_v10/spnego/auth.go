package spnego

import (
	"errors"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
)

// AuthType represents the authentication type
type AuthType int

const (
	AuthTypeNTLM AuthType = iota
	AuthTypeKerberos
)

// AuthContext holds the state for an authentication session
type AuthContext struct {
	Type        AuthType
	Domain      string
	Username    string
	Password    string
	Workstation string
	UseUnicode  bool

	// NTLM specific fields
	NTLMChallenge *ntlm.ChallengeMessage
}

// NewAuthContext creates a new authentication context
func NewAuthContext(authType AuthType, domain, username, password, workstation string, useUnicode bool) *AuthContext {
	return &AuthContext{
		Type:        authType,
		Domain:      domain,
		Username:    username,
		Password:    password,
		Workstation: workstation,
		UseUnicode:  useUnicode,
	}
}

// ProcessChallengeToken processes the server's challenge token and prepares the authenticate token
func (ctx *AuthContext) ProcessChallengeToken(token []byte) ([]byte, error) {
	// Parse the SPNEGO token
	resp, err := ParseNegTokenResp(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SPNEGO token: %v", err)
	}

	// Check if the server accepted our mechanism
	if resp.NegState == Reject {
		return nil, errors.New("server rejected authentication")
	}

	// Extract the inner token
	innerToken, err := ExtractNTLMToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to extract inner token: %v", err)
	}

	switch ctx.Type {
	case AuthTypeNTLM:
		// Parse the NTLM CHALLENGE message
		challenge, err := ntlm.ParseChallengeMessage(innerToken)
		if err != nil {
			return nil, fmt.Errorf("failed to parse NTLM CHALLENGE message: %v", err)
		}

		// Store the challenge for later use
		ctx.NTLMChallenge = challenge

		// Create NTLM AUTHENTICATE message
		ntlmAuth, err := ntlm.CreateAuthenticateMessage(challenge, ctx.Username, ctx.Password, ctx.Domain, ctx.Workstation)
		if err != nil {
			return nil, fmt.Errorf("failed to create NTLM AUTHENTICATE message: %v", err)
		}

		// Wrap in SPNEGO
		return CreateNegTokenInit(ntlmAuth)

	case AuthTypeKerberos:
		return nil, errors.New("kerberos authentication is not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported authentication type: %v", ctx.Type)
	}
}

// PrepareSessionSetupRequest prepares the SMB session setup request with SPNEGO token
func PrepareSessionSetupRequest(token []byte, useUnicode bool) []byte {
	if useUnicode {
		return utf16.EncodeUTF16LE(string(token))
	} else {
		return token
	}
}
