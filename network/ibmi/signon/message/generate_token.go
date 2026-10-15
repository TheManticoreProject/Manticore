package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Token-type bytes of CP 0x1116.
const (
	TokenTypeSingleUse            byte   = 0xF1
	TokenTypeMultipleUse          byte   = 0xF2
	TokenTypeMultipleUseRenewable byte   = 0xF3
	outputKindProfileToken        byte   = 0x01
	MaxTokenTimeoutSeconds        uint32 = 3600
)

// GenerateTokenRequest is request 0x7007: issue a profile token
// authenticated by a password substitute. Its frame contains the
// substitute: never log it, and wipe the frame's Variable slice after
// sending.
type GenerateTokenRequest struct {
	// CredentialType is the input authentication kind (password
	// substitute types only).
	CredentialType byte
	Substitute     []byte
	Profile        [10]byte
	// TokenType is TokenTypeSingleUse, TokenTypeMultipleUse or
	// TokenTypeMultipleUseRenewable.
	TokenType byte
	// TimeoutSeconds is the token lifetime, 1 to 3600.
	TimeoutSeconds uint32
}

// Frame encodes the request: template = input kind and output kind 0x01
// (profile token), then CP 0x1116 token type, CP 0x1117 timeout, CP 0x1105
// substitute and CP 0x1104 profile.
func (r GenerateTokenRequest) Frame() (hostserver.Frame, error) {
	if r.CredentialType == CredentialTypeProfileToken {
		return hostserver.Frame{}, fmt.Errorf("message: token generation needs a password substitute")
	}
	if r.TokenType < TokenTypeSingleUse || r.TokenType > TokenTypeMultipleUseRenewable {
		return hostserver.Frame{}, fmt.Errorf("message: token type 0x%02X undefined", r.TokenType)
	}
	if r.TimeoutSeconds < 1 || r.TimeoutSeconds > MaxTokenTimeoutSeconds {
		return hostserver.Frame{}, fmt.Errorf("message: token timeout %d s outside 1..%d", r.TimeoutSeconds, MaxTokenTimeoutSeconds)
	}
	cred, err := credentialFields(r.CredentialType, r.Substitute, r.Profile)
	if err != nil {
		return hostserver.Frame{}, err
	}
	fields := append([]hostserver.Field{
		{CodePoint: CPTokenType, Value: []byte{r.TokenType}},
		{CodePoint: CPTokenTimeout, Value: u32(r.TimeoutSeconds)},
	}, cred...)
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDSignon, RequestID: RequestGenerateToken},
		Template: []byte{r.CredentialType, outputKindProfileToken},
		Variable: appendFields(fields...),
	}, nil
}

// generateTokenMinBody is the return code, six opaque bytes and the token.
const generateTokenMinBody = 4 + 6 + 32

// GenerateTokenReply is reply 0xF007.
type GenerateTokenReply struct {
	ReturnCode uint32
	// Prefix holds body bytes 4..9 (frame offsets 24..29), whose meaning
	// is not established; kept opaque for capture comparison.
	Prefix [6]byte
	// Token is the 32-byte profile token, set only when ReturnCode is 0.
	// Wipe it once transferred to a credential.
	Token []byte
}

// ParseGenerateTokenReply decodes reply 0xF007: a 4-byte return code and,
// on success, the token at frame offsets 30..61. A nonzero return code never
// yields a token. Bytes after the token are not interpreted.
func ParseGenerateTokenReply(f hostserver.Frame) (*GenerateTokenReply, error) {
	b := body(f)
	if len(b) < 4 {
		return nil, &hostserver.MalformedError{What: "0xF007 reply", Offset: hostserver.HeaderLength, Reason: fmt.Sprintf("%d-byte body cannot hold a return code", len(b))}
	}
	r := &GenerateTokenReply{ReturnCode: binary.BigEndian.Uint32(b)}
	if r.ReturnCode != 0 {
		return r, nil
	}
	if len(b) < generateTokenMinBody {
		return nil, &hostserver.MalformedError{What: "0xF007 reply", Offset: hostserver.HeaderLength, Reason: fmt.Sprintf("%d-byte body, need at least %d", len(b), generateTokenMinBody)}
	}
	copy(r.Prefix[:], b[4:10])
	r.Token = append([]byte(nil), b[10:42]...)
	clear(b)
	return r, nil
}

func body(f hostserver.Frame) []byte {
	b := make([]byte, 0, len(f.Template)+len(f.Variable))
	return append(append(b, f.Template...), f.Variable...)
}
