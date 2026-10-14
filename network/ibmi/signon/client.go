package signon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/auth"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon/message"
)

// DefaultCleanupTimeout bounds the best-effort 0x7006 end request when
// Options.CleanupTimeout is 0.
const DefaultCleanupTimeout = 5 * time.Second

// Options configures Validate.
type Options struct {
	// Dialer opens the Sign-on connection; nil uses a zero Dialer, which
	// requires the system trust store for TLS endpoints.
	Dialer *hostserver.Dialer
	// Rand is the client-seed source; nil uses crypto/rand.
	Rand io.Reader
	// CleanupTimeout bounds the end-server request; 0 means
	// DefaultCleanupTimeout. It runs on a fresh context so a cancelled
	// caller context still ends the server job.
	CleanupTimeout time.Duration
}

// Validate authenticates cred once against the Sign-on server at ep and
// returns the negotiated metadata. The connection is always closed before
// Validate returns. A credential rejected by the server yields an *Error
// matching ErrAuthenticationRejected; Validate never retries.
func Validate(ctx context.Context, ep ibminet.Endpoint, cred *ibminet.PasswordCredential, opts Options) (*Info, error) {
	if ep.Service != ibminet.ServiceSignon {
		return nil, fmt.Errorf("signon: endpoint is for %s, not %s", ep.Service, ibminet.ServiceSignon)
	}
	if cred == nil {
		return nil, errors.New("signon: nil credential")
	}
	if opts.CleanupTimeout < 0 {
		return nil, fmt.Errorf("signon: negative cleanup timeout %s", opts.CleanupTimeout)
	}
	cleanup := opts.CleanupTimeout
	if cleanup == 0 {
		cleanup = DefaultCleanupTimeout
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

	info, err := exchange(ctx, conn, cred, clientSeed)

	// End the server job while the stream is still synchronized, on success
	// and on a clean rejection alike. A poisoned Conn, or one closed after a
	// malformed reply, is skipped.
	if conn.Err() == nil {
		endCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cleanup)
		_, _ = conn.Send(endCtx, message.EndServerFrame())
		cancel()
	}
	if err != nil {
		return nil, err
	}
	return info, nil
}

var expectAttributes = hostserver.Expect{
	ServerID: hostserver.ServerIDSignon, ReplyIDs: []uint16{message.ReplyExchangeAttributes},
}

var expectInfo = hostserver.Expect{
	ServerID: hostserver.ServerIDSignon, ReplyIDs: []uint16{message.ReplyRetrieveInfo},
}

func exchange(ctx context.Context, conn *hostserver.Conn, cred *ibminet.PasswordCredential, clientSeed [auth.SeedLength]byte) (*Info, error) {
	f, _, err := conn.Exchange(ctx, message.ExchangeAttributesRequest{ClientSeed: &clientSeed}.Frame(), expectAttributes)
	if err != nil {
		return nil, fmt.Errorf("signon: exchange attributes: %w", err)
	}
	attrs, err := message.ParseExchangeAttributesReply(f)
	if err != nil {
		conn.Close() // malformed reply: server state unknown, skip 0x7006
		return nil, fmt.Errorf("signon: exchange attributes: %w", err)
	}
	if attrs.ReturnCode != 0 {
		return nil, &Error{Stage: "exchange attributes", Code: ReturnCode(attrs.ReturnCode)}
	}

	info := &Info{ServerVersion: attrs.ServerVersion, ServerLevel: attrs.ServerLevel}
	if vrm, err := decodeVRM(attrs.ServerVersion); err == nil {
		info.VRM = vrm
	} else {
		info.Warnings = append(info.Warnings, fmt.Sprintf("server version 0x%08X does not decode as a VRM", attrs.ServerVersion))
	}
	if attrs.HasPasswordLevel {
		info.PasswordLevel, info.PasswordLevelReported = attrs.PasswordLevel, true
		// Fails for level 4+ before any credential material is derived.
		if info.PasswordFamily, err = auth.FamilyForLevel(attrs.PasswordLevel); err != nil {
			return nil, fmt.Errorf("signon: %w", err)
		}
	} else {
		info.PasswordFamily = auth.FamilyDES
		info.Warnings = append(info.Warnings, "server omitted the password level (CP 0x1119); assumed levels 0/1 (DES)")
	}

	sub, err := auth.Derive(cred, info.PasswordFamily, clientSeed, attrs.ServerSeed)
	if err != nil {
		return nil, fmt.Errorf("signon: %w", err)
	}
	defer sub.Wipe()
	req, err := message.RetrieveInfoRequest{CredentialType: sub.CredentialType, Substitute: sub.Value, Profile: sub.Profile}.Frame()
	if err != nil {
		return nil, fmt.Errorf("signon: %w", err)
	}
	f, _, err = conn.Exchange(ctx, req, expectInfo)
	clear(req.Variable) // holds the substitute and profile
	if err != nil {
		return nil, fmt.Errorf("signon: retrieve information: %w", err)
	}
	ri, err := message.ParseRetrieveInfoReply(f)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("signon: retrieve information: %w", err)
	}
	if ri.ReturnCode != 0 {
		return nil, &Error{Stage: "retrieve information", Code: ReturnCode(ri.ReturnCode)}
	}
	if !ri.HasServerCCSID {
		conn.Close()
		return nil, fmt.Errorf("signon: retrieve information: %w", &hostserver.MalformedError{What: "reply", Reason: "system CCSID (CP 0x1114) is missing"})
	}
	info.SystemCCSID = ri.ServerCCSID
	return info, nil
}
