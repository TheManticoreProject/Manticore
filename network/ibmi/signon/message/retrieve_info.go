package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Credential types carried in the one-byte 0x7004 template.
const (
	CredentialTypeDES  byte = 0x01
	CredentialTypeSHA1 byte = 0x03
)

// RetrieveInfoRequest is request 0x7004: it validates a password substitute.
// The frame contains the substitute; never log it, and wipe the frame's
// Variable slice after sending.
type RetrieveInfoRequest struct {
	CredentialType byte
	Substitute     []byte
	// Profile is the 10-byte EBCDIC sign-on profile.
	Profile [10]byte
}

// Frame encodes the request: template = credential type, then CP 0x1113
// client CCSID 13488, CP 0x1105 substitute and CP 0x1104 profile.
func (r RetrieveInfoRequest) Frame() (hostserver.Frame, error) {
	want := map[byte]int{CredentialTypeDES: 8, CredentialTypeSHA1: 20}[r.CredentialType]
	if want == 0 {
		return hostserver.Frame{}, fmt.Errorf("message: unsupported credential type 0x%02X", r.CredentialType)
	}
	if len(r.Substitute) != want {
		return hostserver.Frame{}, fmt.Errorf("message: credential type 0x%02X needs a %d-byte substitute, got %d", r.CredentialType, want, len(r.Substitute))
	}
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDSignon, RequestID: RequestRetrieveInfo},
		Template: []byte{r.CredentialType},
		Variable: appendFields(
			hostserver.Field{CodePoint: CPClientCCSID, Value: u32(ClientCCSID)},
			hostserver.Field{CodePoint: CPPasswordSubstitute, Value: r.Substitute},
			hostserver.Field{CodePoint: CPUserID, Value: r.Profile[:]},
		),
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
