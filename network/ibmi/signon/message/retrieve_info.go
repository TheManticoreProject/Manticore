package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Credential types carried in the one-byte 0x7004 template.
const (
	CredentialTypeDES          byte = 0x01
	CredentialTypeProfileToken byte = 0x02 // CP 0x1115, no profile
	CredentialTypeSHA1         byte = 0x03
	CredentialTypeSHA512       byte = 0x07 // password level 4
)

// credentialLengths is the exact material length of each credential type.
var credentialLengths = map[byte]int{
	CredentialTypeDES: 8, CredentialTypeProfileToken: 32, CredentialTypeSHA1: 20, CredentialTypeSHA512: 64,
}

// credentialFields encodes the authentication material: a password
// substitute in CP 0x1105 followed by the profile in CP 0x1104, or a
// profile token in CP 0x1115 with no profile.
func credentialFields(credType byte, material []byte, profile [10]byte) ([]hostserver.Field, error) {
	want := credentialLengths[credType]
	if want == 0 {
		return nil, fmt.Errorf("message: unsupported credential type 0x%02X", credType)
	}
	if len(material) != want {
		return nil, fmt.Errorf("message: credential type 0x%02X needs %d bytes, got %d", credType, want, len(material))
	}
	if credType == CredentialTypeProfileToken {
		return []hostserver.Field{{CodePoint: CPAuthToken, Value: material}}, nil
	}
	return []hostserver.Field{
		{CodePoint: CPPasswordSubstitute, Value: material},
		{CodePoint: CPUserID, Value: profile[:]},
	}, nil
}

// RetrieveInfoRequest is request 0x7004: it validates a password substitute
// or a profile token.
// The frame contains the substitute; never log it, and wipe the frame's
// Variable slice after sending.
type RetrieveInfoRequest struct {
	CredentialType byte
	Substitute     []byte
	// Profile is the 10-byte EBCDIC sign-on profile.
	Profile [10]byte
}

// Frame encodes the request: template = credential type, then CP 0x1113
// client CCSID 13488 and the credential (CP 0x1105 substitute and CP 0x1104
// profile, or CP 0x1115 profile token).
func (r RetrieveInfoRequest) Frame() (hostserver.Frame, error) {
	cred, err := credentialFields(r.CredentialType, r.Substitute, r.Profile)
	if err != nil {
		return hostserver.Frame{}, err
	}
	fields := append([]hostserver.Field{{CodePoint: CPClientCCSID, Value: u32(ClientCCSID)}}, cred...)
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDSignon, RequestID: RequestRetrieveInfo},
		Template: []byte{r.CredentialType},
		Variable: appendFields(fields...),
	}, nil
}

// RetrieveInfoReply is reply 0xF004.
type RetrieveInfoReply struct {
	ReturnCode uint32
	// ServerCCSID is CP 0x1114; HasServerCCSID is false when absent.
	ServerCCSID    uint32
	HasServerCCSID bool
	// Fields holds every reply field. It includes the effective profile
	// (CP 0x1104), so it is not exposed beyond the Sign-on package.
	Fields hostserver.Fields
}

// ParseRetrieveInfoReply decodes reply 0xF004. A nonzero return code is
// returned for the caller to classify.
func ParseRetrieveInfoReply(f hostserver.Frame) (*RetrieveInfoReply, error) {
	rc, fields, err := replyBody(f, CPServerCCSID, CPUserID)
	if err != nil {
		return nil, err
	}
	r := &RetrieveInfoReply{ReturnCode: rc, Fields: fields}
	if v, ok, err := fixedField(fields, CPServerCCSID, 4); err != nil {
		return nil, err
	} else if ok {
		r.ServerCCSID, r.HasServerCCSID = binary.BigEndian.Uint32(v), true
	}
	return r, nil
}
