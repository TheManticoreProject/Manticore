package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// StartAttributes is header byte 4 of 0x7002: ask the server to return its
// job information.
const StartAttributes byte = 0x02

// Credential types carried in the first 0x7002 template byte.
const (
	CredentialTypeDES    byte = 0x01
	CredentialTypeSHA1   byte = 0x03
	CredentialTypeSHA512 byte = 0x07 // password level 4
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
// then CP 0x1105 substitute and CP 0x1104 profile.
func (r StartServerRequest) Frame() (hostserver.Frame, error) {
	want := map[byte]int{CredentialTypeDES: 8, CredentialTypeSHA1: 20, CredentialTypeSHA512: 64}[r.CredentialType]
	if want == 0 {
		return hostserver.Frame{}, fmt.Errorf("message: unsupported credential type 0x%02X", r.CredentialType)
	}
	if len(r.Substitute) != want {
		return hostserver.Frame{}, fmt.Errorf("message: credential type 0x%02X needs a %d-byte substitute, got %d", r.CredentialType, want, len(r.Substitute))
	}
	var v []byte
	v, _ = hostserver.AppendField(v, CPPasswordSubstitute, r.Substitute)
	v, _ = hostserver.AppendField(v, CPUserID, r.Profile[:])
	return hostserver.Frame{
		Header: hostserver.Header{
			Attributes: uint16(StartAttributes) << 8,
			ServerID:   hostserver.ServerIDRemoteCommand,
			RequestID:  RequestStartServer,
		},
		Template: []byte{r.CredentialType, 0x01},
		Variable: v,
	}, nil
}

// StartServerReply is reply 0xF002.
type StartServerReply struct {
	ReturnCode uint32
	// JobIdentity is the raw CP 0x111F value: a 4-byte opaque prefix, then
	// the server job name in the job CCSID. HasJobIdentity is false when
	// absent.
	JobIdentity    []byte
	HasJobIdentity bool
	Fields         hostserver.Fields
}

// ParseStartServerReply decodes reply 0xF002: a 4-byte return code, then
// LL/CP fields.
func ParseStartServerReply(f hostserver.Frame) (*StartServerReply, error) {
	b := body(f)
	if len(b) < 4 {
		return nil, short("0xF002 reply", len(b), 4)
	}
	fields, err := hostserver.ParseFields(b[4:], hostserver.FieldPolicy{
		MaxFields: maxReplyFields, Singletons: map[uint16]bool{CPJobIdentity: true, CPUserID: true},
	})
	if err != nil {
		return nil, err
	}
	r := &StartServerReply{ReturnCode: binary.BigEndian.Uint32(b), Fields: fields}
	if v, ok := fields.Get(CPJobIdentity); ok {
		if len(v) < 4 {
			return nil, &hostserver.MalformedError{What: "field", Reason: fmt.Sprintf("job identity (CP 0x111F) has %d bytes, need at least 4", len(v))}
		}
		r.JobIdentity, r.HasJobIdentity = append([]byte(nil), v...), true
	}
	return r, nil
}
