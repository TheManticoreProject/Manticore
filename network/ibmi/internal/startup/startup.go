// Package startup implements the authentication every IBM i host server
// except Sign-on performs on a new connection: exchange seeds
// (0x7001/0xF001) then start server (0x7002/0xF002), with the password
// substitute or profile token from network/ibmi/internal/auth. The frames
// differ between servers only in the header server ID.
//
// Layouts follow the JTOpen 4.8 generation that the V5R4 baseline targets,
// as recorded in iSeriesNavigatorReverse docs/04-host-server-core-protocol.md
// and the golden frames in docs/10-v5r4-compatibility-and-golden-vectors.md.
package startup

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/auth"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon"
)

// Request and reply IDs.
const (
	RequestExchangeSeeds uint16 = 0x7001
	ReplyExchangeSeeds   uint16 = 0xF001
	RequestStartServer   uint16 = 0x7002
	ReplyStartServer     uint16 = 0xF002
)

// Code points used by the startup exchanges.
const (
	CPUserID             uint16 = 0x1104
	CPPasswordSubstitute uint16 = 0x1105
	CPJobIdentity        uint16 = 0x111F
	CPAuthToken          uint16 = 0x1115
)

// Header byte 4 of 0x7001 advertises the client's password capabilities:
// the JTOpen 4.8 client sends SeedAttributes (SHA-1 capable); current JTOpen
// sends SeedAttributesLevel4, which also advertises the password level 4
// algorithm (iSeriesNavigatorReverse docs/04).
const (
	SeedAttributes       byte = 0x01
	SeedAttributesLevel4 byte = 0x03
)

// StartAttributes is header byte 4 of 0x7002: ask the server to return its
// job information.
const StartAttributes byte = 0x02

// Credential types carried in the first 0x7002 template byte.
const (
	CredentialTypeDES          byte = 0x01
	CredentialTypeProfileToken byte = 0x02 // CP 0x1115, no profile
	CredentialTypeSHA1         byte = 0x03
	CredentialTypeSHA512       byte = 0x07 // password level 4
)

// maxReplyFields bounds the LL/CP fields accepted in one startup reply.
const maxReplyFields = 64

// Body returns a reply's bytes after the header, whatever its declared
// template length.
func Body(f hostserver.Frame) []byte {
	b := make([]byte, 0, len(f.Template)+len(f.Variable))
	return append(append(b, f.Template...), f.Variable...)
}

// Short reports a reply body shorter than want bytes.
func Short(what string, have, want int) error {
	return &hostserver.MalformedError{What: what, Offset: hostserver.HeaderLength, Reason: fmt.Sprintf("%d-byte body, need at least %d", have, want)}
}

// ExchangeSeedsFrame encodes request 0x7001 for serverID: the 8-byte client
// seed is the whole template.
func ExchangeSeedsFrame(serverID uint16, clientSeed [8]byte, attributes byte) hostserver.Frame {
	return hostserver.Frame{
		Header: hostserver.Header{
			Attributes: uint16(attributes) << 8,
			ServerID:   serverID,
			RequestID:  RequestExchangeSeeds,
		},
		Template: clientSeed[:],
	}
}

// ExchangeSeedsReply is reply 0xF001.
type ExchangeSeedsReply struct {
	ReturnCode uint32
	ServerSeed [8]byte
	// ServerAttributes is header byte 5 (the server's password-level
	// attributes on V5R1 and later). It is informational.
	ServerAttributes byte
	// Fields holds any optional LL/CP fields after the seed.
	Fields hostserver.Fields
}

// ParseExchangeSeedsReply decodes reply 0xF001: a 4-byte return code, then
// with a zero return code the 8-byte server seed and optional LL/CP fields.
func ParseExchangeSeedsReply(f hostserver.Frame) (*ExchangeSeedsReply, error) {
	b := Body(f)
	if len(b) < 4 {
		return nil, Short("0xF001 reply", len(b), 4)
	}
	r := &ExchangeSeedsReply{ReturnCode: binary.BigEndian.Uint32(b), ServerAttributes: byte(f.Header.Attributes)}
	if r.ReturnCode != 0 {
		return r, nil
	}
	if len(b) < 12 {
		return nil, Short("0xF001 reply", len(b), 12)
	}
	copy(r.ServerSeed[:], b[4:12])
	fields, err := hostserver.ParseFields(b[12:], hostserver.FieldPolicy{MaxFields: maxReplyFields, Singletons: map[uint16]bool{}})
	if err != nil {
		return nil, err
	}
	r.Fields = fields
	return r, nil
}

// StartServerRequest is request 0x7002. Its frame contains the substitute:
// never log it, and wipe the frame's Variable slice after sending.
type StartServerRequest struct {
	ServerID       uint16
	CredentialType byte
	Substitute     []byte
	// Profile is the 10-byte EBCDIC sign-on profile.
	Profile [10]byte
}

// Frame encodes the request: template = credential type and send-reply 1,
// then CP 0x1105 substitute and CP 0x1104 profile, or CP 0x1115 profile
// token alone.
func (r StartServerRequest) Frame() (hostserver.Frame, error) {
	want := map[byte]int{CredentialTypeDES: 8, CredentialTypeProfileToken: 32, CredentialTypeSHA1: 20, CredentialTypeSHA512: 64}[r.CredentialType]
	if want == 0 {
		return hostserver.Frame{}, fmt.Errorf("message: unsupported credential type 0x%02X", r.CredentialType)
	}
	if len(r.Substitute) != want {
		return hostserver.Frame{}, fmt.Errorf("message: credential type 0x%02X needs %d bytes, got %d", r.CredentialType, want, len(r.Substitute))
	}
	var v []byte
	if r.CredentialType == CredentialTypeProfileToken {
		// A profile token identifies its profile; no CP 0x1104 is sent.
		v, _ = hostserver.AppendField(v, CPAuthToken, r.Substitute)
	} else {
		v, _ = hostserver.AppendField(v, CPPasswordSubstitute, r.Substitute)
		v, _ = hostserver.AppendField(v, CPUserID, r.Profile[:])
	}
	return hostserver.Frame{
		Header: hostserver.Header{
			Attributes: uint16(StartAttributes) << 8,
			ServerID:   r.ServerID,
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
	b := Body(f)
	if len(b) < 4 {
		return nil, Short("0xF002 reply", len(b), 4)
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

// Server names one host server for Authenticate.
type Server struct {
	ID uint16
	// Prefix starts error messages ("rmtcmd"); Stage names the server in
	// signon.Error stages ("remote command").
	Prefix, Stage string
}

// Authenticate starts the server job on conn for cred: a password, with a
// fresh seed pair and substitute, or a profile token. family is the
// password family learned from Sign-on; it is ignored for a profile token.
// A rejected credential returns a *signon.Error matching
// signon.ErrAuthenticationRejected and is never retried. The caller owns
// conn and closes it on error.
func Authenticate(ctx context.Context, conn *hostserver.Conn, srv Server, cred ibminet.Credential, family ibminet.PasswordFamily, rnd io.Reader) (*StartServerReply, error) {
	if cred == nil {
		return nil, fmt.Errorf("%s: nil credential", srv.Prefix)
	}
	clientSeed, err := auth.NewSeed(rnd)
	if err != nil {
		return nil, err
	}
	attributes := SeedAttributes // the V5R4 golden value
	if _, isPassword := cred.(*ibminet.PasswordCredential); isPassword && family == auth.FamilySHA512 {
		attributes = SeedAttributesLevel4
	}
	f, _, err := conn.Exchange(ctx, ExchangeSeedsFrame(srv.ID, clientSeed, attributes),
		hostserver.Expect{ServerID: srv.ID, ReplyIDs: []uint16{ReplyExchangeSeeds}})
	if err != nil {
		return nil, fmt.Errorf("%s: exchange seeds: %w", srv.Prefix, err)
	}
	seeds, err := ParseExchangeSeedsReply(f)
	if err != nil {
		return nil, fmt.Errorf("%s: exchange seeds: %w", srv.Prefix, err)
	}
	if seeds.ReturnCode != 0 {
		return nil, &signon.Error{Stage: srv.Stage + " exchange seeds", Code: signon.ReturnCode(seeds.ReturnCode)}
	}

	sub, err := auth.Derive(cred, family, clientSeed, seeds.ServerSeed)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", srv.Prefix, err)
	}
	req, err := StartServerRequest{ServerID: srv.ID, CredentialType: sub.CredentialType, Substitute: sub.Value, Profile: sub.Profile}.Frame()
	sub.Wipe()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", srv.Prefix, err)
	}
	f, _, err = conn.Exchange(ctx, req, hostserver.Expect{ServerID: srv.ID, ReplyIDs: []uint16{ReplyStartServer}})
	clear(req.Variable) // holds the substitute and profile
	if err != nil {
		return nil, fmt.Errorf("%s: start server: %w", srv.Prefix, err)
	}
	started, err := ParseStartServerReply(f)
	if err != nil {
		return nil, fmt.Errorf("%s: start server: %w", srv.Prefix, err)
	}
	if started.ReturnCode != 0 {
		return nil, &signon.Error{Stage: srv.Stage + " start server", Code: signon.ReturnCode(started.ReturnCode)}
	}
	return started, nil
}
