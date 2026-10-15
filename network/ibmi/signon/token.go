package signon

import (
	"context"
	"errors"
	"fmt"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/auth"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon/message"
)

// TokenRequest describes a profile token to issue.
type TokenRequest struct {
	// Type defaults to single use.
	Type ibminet.ProfileTokenType
	// Timeout is the token lifetime: 1 second to 1 hour, in whole seconds.
	// Zero means 5 minutes.
	Timeout time.Duration
}

// DefaultTokenTimeout is the lifetime used when TokenRequest.Timeout is 0.
const DefaultTokenTimeout = 5 * time.Minute

var expectGenerateToken = hostserver.Expect{
	ServerID: hostserver.ServerIDSignon, ReplyIDs: []uint16{message.ReplyGenerateToken},
}

// GenerateProfileToken authenticates cred with the Sign-on server at ep and
// asks it to issue a profile token (request 0x7007), on one short-lived
// connection.
//
// Issuing a token is visible on the server: it can update the profile's
// last-used date, reset its invalid sign-on count and write an audit entry.
// It is never retried; a rejected password returns an *Error matching
// ErrAuthenticationRejected. The returned token is a bearer secret bound to
// that system: Destroy it when done, and send it only over TLS outside
// isolated labs.
func GenerateProfileToken(ctx context.Context, ep ibminet.Endpoint, cred *ibminet.PasswordCredential, req TokenRequest, opts Options) (*ibminet.ProfileToken, error) {
	if ep.Service != ibminet.ServiceSignon {
		return nil, fmt.Errorf("signon: endpoint is for %s, not %s", ep.Service, ibminet.ServiceSignon)
	}
	if cred == nil {
		return nil, errors.New("signon: nil credential")
	}
	typ := req.Type
	if typ == 0 {
		typ = auth.TokenSingleUse
	}
	if !typ.Valid() {
		return nil, fmt.Errorf("signon: undefined token type %s", typ)
	}
	timeout := req.Timeout
	if timeout == 0 {
		timeout = DefaultTokenTimeout
	}
	if timeout < time.Second || timeout > time.Duration(message.MaxTokenTimeoutSeconds)*time.Second || timeout%time.Second != 0 {
		return nil, fmt.Errorf("signon: token timeout %s must be whole seconds from 1s to 1h", timeout)
	}
	if opts.CleanupTimeout < 0 {
		return nil, fmt.Errorf("signon: negative cleanup timeout %s", opts.CleanupTimeout)
	}
	dialer := opts.Dialer
	if dialer == nil {
		dialer = &hostserver.Dialer{}
	}
	clientSeed, err := auth.NewSeed(opts.Rand)
	if err != nil {
		return nil, err
	}
	conn, err := dialer.DialEndpoint(ctx, ep)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	token, err := generate(ctx, conn, cred, clientSeed, typ, timeout)
	endServer(ctx, conn, opts)
	return token, err
}

func generate(ctx context.Context, conn *hostserver.Conn, cred *ibminet.PasswordCredential, clientSeed [auth.SeedLength]byte, typ ibminet.ProfileTokenType, timeout time.Duration) (*ibminet.ProfileToken, error) {
	info, serverSeed, err := exchangeAttributes(ctx, conn, cred, clientSeed)
	if err != nil {
		return nil, err
	}
	sub, err := auth.Derive(cred, info.PasswordFamily, clientSeed, serverSeed)
	if err != nil {
		return nil, fmt.Errorf("signon: %w", err)
	}
	defer sub.Wipe()
	issuedAt := time.Now()
	frame, err := message.GenerateTokenRequest{
		CredentialType: sub.CredentialType, Substitute: sub.Value, Profile: sub.Profile,
		TokenType:      message.TokenTypeSingleUse + byte(typ-auth.TokenSingleUse),
		TimeoutSeconds: uint32(timeout / time.Second),
	}.Frame()
	if err != nil {
		return nil, fmt.Errorf("signon: %w", err)
	}
	f, _, err := conn.Exchange(ctx, frame, expectGenerateToken)
	clear(frame.Variable)
	if err != nil {
		return nil, fmt.Errorf("signon: generate profile token: %w", err)
	}
	reply, err := message.ParseGenerateTokenReply(f)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("signon: generate profile token: %w", err)
	}
	if reply.ReturnCode != 0 {
		return nil, &Error{Stage: "generate profile token", Code: ReturnCode(reply.ReturnCode)}
	}
	defer clear(reply.Token)
	return auth.NewProfileToken(reply.Token, typ, issuedAt.Add(timeout))
}
