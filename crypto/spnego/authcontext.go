package spnego

import (
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
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
	NTLMChallenge *challenge.ChallengeMessage

	// SessionKey is the session key derived during the most recent successful
	// CreateAuthenticateTokenFromChallengeToken call. It is not transmitted on the
	// wire; callers use it as the MAC key for SMB message signing. It is nil until
	// authentication has produced a key (and for auth paths that do not derive one).
	SessionKey []byte
}

// GetSessionKey returns the session key derived during authentication, or nil if
// none has been derived yet.
//
// Returns:
//   - []byte: The session key (MAC key for SMB signing), or nil
func (ctx *AuthContext) GetSessionKey() []byte {
	return ctx.SessionKey
}

// NewAuthContext creates a new authentication context
// Parameters:
//   - authType: The type of authentication to use (NTLM or Kerberos)
//   - domain: The domain name for authentication
//   - username: The username to authenticate with
//   - password: The password for authentication
//   - workstation: The name of the client workstation
//   - useUnicode: Whether to use Unicode encoding
//
// Returns:
//   - *AuthContext: A new authentication context initialized with the provided parameters
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

// PrepareSessionSetupRequest prepares the SMB session setup request with SPNEGO token
// Parameters:
//   - token: The SPNEGO token bytes to prepare
//   - useUnicode: Whether to encode the token in UTF-16LE
//
// Returns:
//   - []byte: The prepared token, encoded in UTF-16LE if useUnicode is true
func PrepareSessionSetupRequest(token []byte, useUnicode bool) []byte {
	if useUnicode {
		return utf16.EncodeUTF16LE(string(token))
	} else {
		return token
	}
}

// GetAuthType returns the authentication type
// Returns:
//   - AuthType: The authentication type (NTLM or Kerberos) for this context
func (ctx *AuthContext) GetAuthType() AuthType {
	return ctx.Type
}

// SetAuthType sets the authentication type
// Parameters:
//   - authType: The authentication type (NTLM or Kerberos) to set
func (ctx *AuthContext) SetAuthType(authType AuthType) {
	ctx.Type = authType
}
