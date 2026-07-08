package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"io"
	"time"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// DefaultNegotiateFlags is a capable client negotiate-flag set ([MS-NRPC] 3.1.4.2) that
// advertises AES/strong-key secure-channel support. Establish forces NegotiateAES on top of
// whatever flags it is given, because SecureChannel implements only the AES cipher suite.
const DefaultNegotiateFlags uint32 = 0x212fffff

// SecureChannelConfig holds the inputs to Establish. Exactly one of Password or NTHash
// supplies the machine-account secret: NTHash (the raw 16-byte NT hash) takes precedence
// when non-nil, otherwise the NT one-way function of Password is used.
type SecureChannelConfig struct {
	// PrimaryName is the DC name (the server principal), e.g. "DC01". Empty sends a NULL
	// PrimaryName, which asks the server to use its own name.
	PrimaryName string
	// ComputerName is the client's NetBIOS computer name (without the trailing '$'), e.g.
	// "WORKSTATION".
	ComputerName string
	// AccountName is the account the channel authenticates as, typically the machine account
	// "COMPUTER$" ([MS-NRPC] 3.1.4.1).
	AccountName string
	// SecureChannelType is the kind of secure channel being set up, e.g.
	// WorkstationSecureChannel or ServerSecureChannel.
	SecureChannelType msnrpc.NETLOGON_SECURE_CHANNEL_TYPE
	// Password is the account cleartext password; ignored when NTHash is set.
	Password string
	// NTHash is the raw 16-byte NT hash of the account secret (pass-the-hash); nil to derive
	// the key material from Password instead.
	NTHash []byte
	// NegotiateFlags is the client negotiate-flag set; zero selects DefaultNegotiateFlags.
	NegotiateFlags uint32
	// Rand is the source of the 8-byte client challenge; nil selects crypto/rand. It exists
	// as a seam for deterministic testing and must be left nil in production.
	Rand io.Reader
}

// SecureChannel is an established Netlogon secure channel ([MS-NRPC] 3.1.4): it holds the
// negotiated AES session key and negotiate flags and maintains the rolling stored credential
// used to compute and verify per-call Netlogon authenticators (3.1.4.5).
//
// A SecureChannel is stateful and not safe for concurrent use: the stored credential
// advances on every NextAuthenticator/VerifyResponseAuthenticator call, and the two must be
// paired around each authenticated request in the order the calls go on the wire.
type SecureChannel struct {
	sessionKey       [16]byte
	negotiateFlags   uint32
	clientStoredCred msnrpc.NETLOGON_CREDENTIAL
	now              func() time.Time
}

// Establish runs the Netlogon secure-channel handshake over the bound RPC connection rpc
// ([MS-NRPC] 3.1.4.1, AES suite): it generates a client challenge, exchanges it for the
// server challenge via NetrServerReqChallenge (opnum 4), derives the AES session key, and
// proves possession of the machine secret via NetrServerAuthenticate3 (opnum 26). It then
// verifies the server's returned credential equals ComputeNetlogonCredentialAES of the
// server challenge before returning the channel; a mismatch means the server did not prove
// knowledge of the shared secret and the channel is rejected.
//
// rpc must already be bound to the Netlogon interface (an anonymous/unauthenticated binding
// is sufficient for the handshake itself).
func Establish(rpc ndr.Invoker, cfg SecureChannelConfig) (*SecureChannel, error) {
	src := cfg.Rand
	if src == nil {
		src = rand.Reader
	}
	var clientChallenge msnrpc.NETLOGON_CREDENTIAL
	if _, err := io.ReadFull(src, clientChallenge[:]); err != nil {
		return nil, fmt.Errorf("netlogon secure channel: generate client challenge: %w", err)
	}

	serverChallenge, status, err := NetrServerReqChallenge(rpc, cfg.PrimaryName, cfg.ComputerName, clientChallenge)
	if err != nil {
		return nil, fmt.Errorf("netlogon secure channel: NetrServerReqChallenge: %w", err)
	}
	if status != logon.StatusSuccess {
		return nil, fmt.Errorf("netlogon secure channel: NetrServerReqChallenge: %s", logon.StatusString(status))
	}

	sessionKey := ComputeSessionKeyAES(cfg.Password, cfg.NTHash, clientChallenge, serverChallenge)
	clientCredential := ComputeNetlogonCredentialAES(clientChallenge, sessionKey)

	flags := cfg.NegotiateFlags
	if flags == 0 {
		flags = DefaultNegotiateFlags
	}
	flags |= logon.NegotiateAES

	var primaryPtr *ndr.WSTR
	if cfg.PrimaryName != "" {
		p := ndr.WSTR(cfg.PrimaryName)
		primaryPtr = &p
	}

	serverCredential, negFlags, _, err := NetrServerAuthenticate3(
		rpc,
		primaryPtr,
		ndr.WSTR(cfg.AccountName),
		cfg.SecureChannelType,
		ndr.WSTR(cfg.ComputerName),
		clientCredential,
		ndr.DWORD(flags),
	)
	if err != nil {
		return nil, fmt.Errorf("netlogon secure channel: NetrServerAuthenticate3: %w", err)
	}

	expectedServer := ComputeNetlogonCredentialAES(serverChallenge, sessionKey)
	if subtle.ConstantTimeCompare(expectedServer[:], serverCredential[:]) != 1 {
		return nil, fmt.Errorf("netlogon secure channel: server credential mismatch (access denied or wrong machine secret)")
	}

	return &SecureChannel{
		sessionKey:       sessionKey,
		negotiateFlags:   uint32(negFlags),
		clientStoredCred: clientCredential,
		now:              time.Now,
	}, nil
}

// SessionKey returns the negotiated 16-byte AES session key.
func (s *SecureChannel) SessionKey() [16]byte { return s.sessionKey }

// NegotiateFlags returns the negotiate flags the server agreed to (its echoed subset).
func (s *SecureChannel) NegotiateFlags() uint32 { return s.negotiateFlags }

// NextAuthenticator computes the client authenticator to send with the next authenticated
// request ([MS-NRPC] 3.1.4.5 step 1): it advances the stored credential by the current
// timestamp and encrypts the result under the session key. It mutates the channel state, so
// each call yields the authenticator for exactly one request and must be followed by a
// VerifyResponseAuthenticator on the server's reply.
func (s *SecureChannel) NextAuthenticator() msnrpc.NETLOGON_AUTHENTICATOR {
	ts := uint32(s.now().Unix())
	s.clientStoredCred = addToCredential(s.clientStoredCred, ts)
	return msnrpc.NETLOGON_AUTHENTICATOR{
		Credential: ComputeNetlogonCredentialAES(s.clientStoredCred, s.sessionKey),
		Timestamp:  ts,
	}
}

// VerifyResponseAuthenticator validates the server's return authenticator ([MS-NRPC]
// 3.1.4.5 step 3): it advances the stored credential by one and checks that encrypting it
// under the session key reproduces the server's credential, in constant time. A mismatch
// means the secure channel is no longer valid and the caller should re-establish it. It must
// be called once per request, paired with the preceding NextAuthenticator.
func (s *SecureChannel) VerifyResponseAuthenticator(server msnrpc.NETLOGON_AUTHENTICATOR) error {
	s.clientStoredCred = addToCredential(s.clientStoredCred, 1)
	expected := ComputeNetlogonCredentialAES(s.clientStoredCred, s.sessionKey)
	if subtle.ConstantTimeCompare(expected[:], server.Credential[:]) != 1 {
		return fmt.Errorf("netlogon secure channel: server authenticator mismatch (channel invalid, re-establish)")
	}
	return nil
}
