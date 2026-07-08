package securechannel

// IDL source: [MS-NRPC] — verified against the protocol's authoritative IDL
// (https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7).

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"io"
	"time"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
	nrpccrypto "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc/crypto"
)

// DefaultNegotiateFlags is a capable client negotiate-flag set ([MS-NRPC] 3.1.4.2) that
// advertises AES support. Establish derives the cipher suite from the flags it is given:
// with NegotiateAES set (as here) it uses the AES suite; with that bit cleared it falls back
// to the legacy strong-key (RC4/DES) suite.
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
	// Whether NegotiateAES is set selects the cipher suite (AES vs legacy strong-key).
	NegotiateFlags uint32
	// Rand is the source of the 8-byte client challenge; nil selects crypto/rand. It exists
	// as a seam for deterministic testing and must be left nil in production.
	Rand io.Reader
}

// SecureChannel is an established Netlogon secure channel ([MS-NRPC] 3.1.4): it holds the
// negotiated session key, the negotiate flags, the cipher suite, and the rolling stored
// credential used to compute and verify per-call Netlogon authenticators (3.1.4.5).
//
// A SecureChannel is stateful and not safe for concurrent use: the stored credential
// advances on every NextAuthenticator/VerifyResponseAuthenticator call, and the two must be
// paired around each authenticated request in the order the calls go on the wire.
type SecureChannel struct {
	sessionKey       [16]byte
	negotiateFlags   uint32
	aes              bool
	clientStoredCred msnrpc.NETLOGON_CREDENTIAL
	now              func() time.Time
}

// Establish runs the Netlogon secure-channel handshake over the bound RPC connection rpc
// ([MS-NRPC] 3.1.4.1): it generates a client challenge, exchanges it for the server challenge
// via NetrServerReqChallenge (opnum 4), derives the session key, and proves possession of the
// machine secret via NetrServerAuthenticate3 (opnum 26). The cipher suite (AES vs legacy
// strong-key) follows the NegotiateAES bit in the negotiate flags. It then verifies the
// server's returned credential equals the computed credential of the server challenge before
// returning the channel; a mismatch means the server did not prove knowledge of the shared
// secret and the channel is rejected.
//
// rpc must already be bound to the Netlogon interface (an anonymous/unauthenticated binding
// is sufficient for the handshake itself).
func Establish(rpc ndr.Invoker, cfg SecureChannelConfig) (*SecureChannel, error) {
	// Validate the machine secret up front so a wrong-length hash or a missing secret fails
	// clearly here rather than deriving a bad session key and failing as ACCESS_DENIED at
	// NetrServerAuthenticate3. An NTHash must be the raw 16 octets; anything else (including
	// a non-nil empty slice) is treated as "not provided" and normalised to nil.
	if len(cfg.NTHash) != 0 && len(cfg.NTHash) != 16 {
		return nil, fmt.Errorf("netlogon secure channel: NTHash must be 16 bytes, got %d", len(cfg.NTHash))
	}
	var ntHash []byte
	if len(cfg.NTHash) == 16 {
		ntHash = cfg.NTHash
	}
	if cfg.Password == "" && ntHash == nil {
		return nil, fmt.Errorf("netlogon secure channel: one of Password or a 16-byte NTHash must be set")
	}

	src := cfg.Rand
	if src == nil {
		src = rand.Reader
	}
	var clientChallenge msnrpc.NETLOGON_CREDENTIAL
	if _, err := io.ReadFull(src, clientChallenge[:]); err != nil {
		return nil, fmt.Errorf("netlogon secure channel: generate client challenge: %w", err)
	}

	serverChallenge, status, err := functions.NetrServerReqChallenge(rpc, cfg.PrimaryName, cfg.ComputerName, clientChallenge)
	if err != nil {
		return nil, fmt.Errorf("netlogon secure channel: NetrServerReqChallenge: %w", err)
	}
	if status != logon.StatusSuccess {
		return nil, fmt.Errorf("netlogon secure channel: NetrServerReqChallenge: %s", logon.StatusString(status))
	}

	flags := cfg.NegotiateFlags
	if flags == 0 {
		flags = DefaultNegotiateFlags
	}
	aes := flags&logon.NegotiateAES != 0

	var sessionKey [16]byte
	if aes {
		sessionKey = nrpccrypto.ComputeSessionKeyAES(cfg.Password, ntHash, clientChallenge, serverChallenge)
	} else {
		sessionKey = nrpccrypto.ComputeSessionKeyStrongKey(cfg.Password, ntHash, clientChallenge, serverChallenge)
	}
	credential := func(in msnrpc.NETLOGON_CREDENTIAL) msnrpc.NETLOGON_CREDENTIAL {
		if aes {
			return nrpccrypto.ComputeNetlogonCredentialAES(in, sessionKey)
		}
		return nrpccrypto.ComputeNetlogonCredential(in, sessionKey)
	}
	clientCredential := credential(clientChallenge)

	var primaryPtr *ndr.WSTR
	if cfg.PrimaryName != "" {
		p := ndr.WSTR(cfg.PrimaryName)
		primaryPtr = &p
	}

	serverCredential, negFlags, _, err := functions.NetrServerAuthenticate3(
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

	expectedServer := credential(serverChallenge)
	if subtle.ConstantTimeCompare(expectedServer[:], serverCredential[:]) != 1 {
		return nil, fmt.Errorf("netlogon secure channel: server credential mismatch (access denied or wrong machine secret)")
	}

	return &SecureChannel{
		sessionKey:       sessionKey,
		negotiateFlags:   uint32(negFlags),
		aes:              aes,
		clientStoredCred: clientCredential,
		now:              time.Now,
	}, nil
}

// SessionKey returns the negotiated 16-byte session key.
func (s *SecureChannel) SessionKey() [16]byte { return s.sessionKey }

// NegotiateFlags returns the negotiate flags the server agreed to (its echoed subset).
func (s *SecureChannel) NegotiateFlags() uint32 { return s.negotiateFlags }

// UsesAES reports whether the AES cipher suite was negotiated; it selects the matching
// MessageSecurity (AES vs RC4) for RPC-level sealing.
func (s *SecureChannel) UsesAES() bool { return s.aes }

// credential computes a Netlogon credential with the channel's cipher suite.
func (s *SecureChannel) credential(input msnrpc.NETLOGON_CREDENTIAL) msnrpc.NETLOGON_CREDENTIAL {
	if s.aes {
		return nrpccrypto.ComputeNetlogonCredentialAES(input, s.sessionKey)
	}
	return nrpccrypto.ComputeNetlogonCredential(input, s.sessionKey)
}

// NextAuthenticator computes the client authenticator to send with the next authenticated
// request ([MS-NRPC] 3.1.4.5 step 1): it advances the stored credential by the current
// timestamp and encrypts the result under the session key. It mutates the channel state, so
// each call yields the authenticator for exactly one request and must be followed by a
// VerifyResponseAuthenticator on the server's reply.
func (s *SecureChannel) NextAuthenticator() msnrpc.NETLOGON_AUTHENTICATOR {
	ts := uint32(s.now().Unix())
	s.clientStoredCred = nrpccrypto.AddToCredential(s.clientStoredCred, ts)
	return msnrpc.NETLOGON_AUTHENTICATOR{
		Credential: s.credential(s.clientStoredCred),
		Timestamp:  ts,
	}
}

// VerifyResponseAuthenticator validates the server's return authenticator ([MS-NRPC]
// 3.1.4.5 step 3): it advances the stored credential by one and checks that encrypting it
// under the session key reproduces the server's credential, in constant time. A mismatch
// means the secure channel is no longer valid and the caller should re-establish it. It must
// be called once per request, paired with the preceding NextAuthenticator.
func (s *SecureChannel) VerifyResponseAuthenticator(server msnrpc.NETLOGON_AUTHENTICATOR) error {
	s.clientStoredCred = nrpccrypto.AddToCredential(s.clientStoredCred, 1)
	expected := s.credential(s.clientStoredCred)
	if subtle.ConstantTimeCompare(expected[:], server.Credential[:]) != 1 {
		return fmt.Errorf("netlogon secure channel: server authenticator mismatch (channel invalid, re-establish)")
	}
	return nil
}
