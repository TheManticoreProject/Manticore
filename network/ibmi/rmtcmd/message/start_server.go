package message

import (
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/startup"
)

// StartAttributes is header byte 4 of 0x7002: ask the server to return its
// job information.
const StartAttributes = startup.StartAttributes

// Credential types carried in the first 0x7002 template byte.
const (
	CredentialTypeDES          = startup.CredentialTypeDES
	CredentialTypeProfileToken = startup.CredentialTypeProfileToken // CP 0x1115, no profile
	CredentialTypeSHA1         = startup.CredentialTypeSHA1
	CredentialTypeSHA512       = startup.CredentialTypeSHA512 // password level 4
)

// StartServerRequest is request 0x7002. Its frame contains the substitute:
// never log it, and wipe the frame's Variable slice after sending.
type StartServerRequest struct {
	CredentialType byte
	Substitute     []byte
	// Profile is the 10-byte EBCDIC sign-on profile.
	Profile [10]byte
}

// Frame encodes the request: template = credential type and send-reply 1,
// then CP 0x1105 substitute and CP 0x1104 profile, or CP 0x1115 profile
// token alone.
func (r StartServerRequest) Frame() (hostserver.Frame, error) {
	return startup.StartServerRequest{
		ServerID: hostserver.ServerIDRemoteCommand, CredentialType: r.CredentialType, Substitute: r.Substitute, Profile: r.Profile,
	}.Frame()
}

// StartServerReply is reply 0xF002.
type StartServerReply = startup.StartServerReply

// ParseStartServerReply decodes reply 0xF002: a 4-byte return code, then
// LL/CP fields.
func ParseStartServerReply(f hostserver.Frame) (*StartServerReply, error) {
	return startup.ParseStartServerReply(f)
}
