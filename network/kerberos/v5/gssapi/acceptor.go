package gssapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credcache/keytab"
	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pac"
)

// The GSS-API acceptor path (RFC 4121 §4.1 / RFC 4120 §3.2). It is the service
// side of context establishment: it consumes the initiator's KRB_AP_REQ GSS
// token, binds the ticket to the configured service principal, decrypts it with
// the service long-term key, validates the
// enclosed authenticator (client identity, clock skew, replay, channel
// bindings), adopts a per-message base key, and — when the initiator asked for
// mutual authentication — emits the KRB_AP_REP GSS token that proves the service
// holds the ticket session key. The resulting SecContext drives the same
// per-message MIC/Wrap layer as the initiator, with the acceptor direction set.

// adTypeIfRelevant / adTypeWin2KPAC are the authorization-data types wrapping the
// PAC inside a Windows ticket (AD-IF-RELEVANT[1] { AD-WIN2K-PAC[128] }).
const (
	adTypeIfRelevant = 1
	adTypeWin2KPAC   = 128
)

// DefaultClockSkew is the maximum difference the acceptor tolerates between the
// authenticator timestamp and its own clock (RFC 4120 §5.3.2 recommends five
// minutes).
const DefaultClockSkew = 5 * time.Minute

// ServiceKey is one candidate long-term key the acceptor can try when decrypting
// the ticket enc-part (key usage 2). It is the explicit-key alternative to a
// keytab.
type ServiceKey struct {
	// EType is the key's Kerberos encryption type (see iana.EType*).
	EType int
	// Key is the raw long-term key bytes.
	Key []byte
}

// TransitedPolicyFunc validates the cross-realm path carried by a service
// ticket. A nil error accepts the path; a non-nil error rejects the ticket.
type TransitedPolicyFunc func(clientRealm, serverRealm string, transited messages.TransitedEncoding) error

// ReplayCache is a minimal in-memory authenticator replay cache (RFC 4120
// §3.2.3): it remembers the (client, ctime, cusec) tuple of every AP-REQ
// authenticator seen within the clock-skew window and rejects a repeat. It is
// safe for concurrent use; a single cache should be shared across all
// AcceptSecContext calls for one service so replays are caught across contexts.
type ReplayCache struct {
	mu      sync.Mutex
	entries map[string]time.Time // tuple -> expiry (ctime + skew)
}

// defaultReplayCache persists replay state across AcceptSecContext calls when a
// caller does not provide a cache. Applications that share a service principal
// across processes still need to provide a shared external replay mechanism.
var defaultReplayCache = NewReplayCache()

// NewReplayCache returns an empty replay cache ready for use.
func NewReplayCache() *ReplayCache {
	return &ReplayCache{entries: make(map[string]time.Time)}
}

// seenBefore records the authenticator tuple and reports whether it was already
// present (a replay). Expired tuples are pruned as a side effect so the cache
// does not grow without bound. now is the acceptor's current time.
func (rc *ReplayCache) seenBefore(tuple string, expiry, now time.Time) bool {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.entries == nil {
		rc.entries = make(map[string]time.Time)
	}
	for k, exp := range rc.entries {
		if now.After(exp) {
			delete(rc.entries, k)
		}
	}
	if _, ok := rc.entries[tuple]; ok {
		return true
	}
	rc.entries[tuple] = expiry
	return false
}

// AcceptOptions configures AcceptSecContext.
type AcceptOptions struct {
	// ServiceName and ServiceRealm identify the service principal this acceptor
	// serves. Both are required and bind accepted tickets to that identity.
	ServiceName  messages.PrincipalName
	ServiceRealm string
	// Keytab supplies candidate service long-term keys (the preferred source):
	// every key whose enctype matches the ticket enc-part is tried at key usage 2.
	Keytab *keytab.Keytab
	// Keys supplies explicit candidate service keys when no keytab is available.
	// They are tried after the keytab keys.
	Keys []ServiceKey
	// UserToUserKey is the target TGT session key used to decrypt a U2U ticket.
	// It is used only when the AP-REQ sets USE-SESSION-KEY; in that case a
	// long-term keytab or Keys entry is never used as a fallback.
	UserToUserKey *ServiceKey
	// ChannelBindings, when non-nil, are the acceptor's channel bindings: the
	// authenticator's 0x8003 Bnd field must equal MD5(ChannelBindings) or the
	// AP-REQ is rejected (GSS_C_BAD_BINDINGS). When nil the initiator's channel
	// bindings are not verified (RFC 4121 §4.1.1 permits the acceptor to ignore
	// them).
	ChannelBindings []byte
	// ClockSkew is the maximum tolerated difference between the authenticator
	// timestamp and the acceptor clock. Zero selects DefaultClockSkew.
	ClockSkew time.Duration
	// ClientAddress is the operating-system reported peer address. It is required
	// when the ticket contains caddr restrictions and ignored for addressless
	// tickets.
	ClientAddress net.IP
	// ReplayCache detects replayed authenticators. When nil the package-wide
	// in-memory cache is used so replay protection persists across calls in this
	// process. Callers serving a principal from multiple processes or machines
	// should pass a cache backed by shared replay state.
	ReplayCache *ReplayCache
	// TransitedPolicy validates cross-realm ticket paths when the service-realm
	// KDC did not set TRANSITED-POLICY-CHECKED. When nil, such unchecked tickets
	// are rejected. If supplied, it is called for every cross-realm ticket so an
	// application can enforce policy even when the KDC set the flag.
	TransitedPolicy TransitedPolicyFunc
	// MintSubkey makes the acceptor generate its own sub-session key, return it in
	// the AP-REP (setting the AcceptorSubkey per-message flag), and key per-message
	// tokens with it. Requires mutual authentication so the initiator learns the
	// key. When false the authenticator subkey (if any) or the ticket session key
	// keys the per-message tokens.
	MintSubkey bool
	// Now overrides the acceptor's notion of the current time (for testing). Zero
	// uses time.Now().
	Now time.Time
	// DisableReplayDetection turns off the per-message receive-side replay window
	// (see InitOptions); replay detection is on by default.
	DisableReplayDetection bool
	// EnforceSequence additionally rejects per-message tokens delivered out of
	// order (see InitOptions); off by default.
	EnforceSequence bool
}

// AcceptSecContext consumes an initiator's KRB_AP_REQ GSS token and establishes
// the acceptor (service) side of the context. It parses the InitialContextToken
// wrapper and the AP-REQ, requires the ticket's server principal to match
// ServiceName@ServiceRealm, decrypts the ticket enc-part with a service long-term
// key (key usage 2, trying keytab and explicit keys), decrypts and validates the
// authenticator with the ticket session key (key usage 11), enforces client
// identity, clock skew, replay and — when requested — channel bindings, and
// adopts a per-message base key. When the AP-REQ set the mutual-required option
// it returns the KRB_AP_REP GSS token (echoing ctime/cusec, key usage 12); the
// output token is nil otherwise.
//
// The returned SecContext exposes the authenticated client identity
// (ClientPrincipal), the extracted PAC (PAC), the negotiated session key
// (SessionKey / SessionEType) and the validated authenticator (Authenticator),
// and drives the per-message MIC/Wrap layer as the acceptor.
func AcceptSecContext(token []byte, opts AcceptOptions) (outputToken []byte, ctx *SecContext, err error) {
	tokID, krbMsg, err := UnwrapToken(token)
	if err != nil {
		return nil, nil, err
	}
	if tokID != TokIDAPReq {
		return nil, nil, fmt.Errorf("gssapi: expected AP-REQ token (01 00), got %02x %02x", tokID[0], tokID[1])
	}

	var apReq messages.APReq
	if _, err := apReq.Unmarshal(krbMsg); err != nil {
		return nil, nil, fmt.Errorf("gssapi: parse AP-REQ: %w", err)
	}
	if len(opts.ServiceName.NameString) == 0 || opts.ServiceRealm == "" {
		return nil, nil, fmt.Errorf("gssapi: expected service principal is required")
	}
	if !principalsEqual(apReq.Ticket.SName, opts.ServiceName) || !strings.EqualFold(apReq.Ticket.Realm, opts.ServiceRealm) {
		return nil, nil, fmt.Errorf("gssapi: ticket service %s@%s does not match expected service %s@%s",
			principalString(apReq.Ticket.SName), apReq.Ticket.Realm, principalString(opts.ServiceName), opts.ServiceRealm)
	}

	// Decrypt the ticket enc-part with a service long-term key (key usage 2) and
	// recover the session key the KDC sealed inside it.
	useSessionKey := apReq.APOptions.At(messages.APOptionUseSessionKey) != 0
	encTkt, err := decryptTicket(&apReq.Ticket, opts, useSessionKey)
	if err != nil {
		return nil, nil, err
	}
	sessionKey := encTkt.Key.KeyValue
	sessionEType := encTkt.Key.KeyType

	// Ticket validity checks (RFC 4120 §3.2.3): reject tickets that are expired,
	// not yet valid, or explicitly marked INVALID (postdated tickets that have not
	// been validated by the KDC).
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	skew := opts.ClockSkew
	if skew <= 0 {
		skew = DefaultClockSkew
	}

	if encTkt.Flags.At(messages.TicketFlagInvalid) != 0 {
		return nil, nil, fmt.Errorf("gssapi: ticket has the INVALID flag set (postdated but not validated)")
	}
	if !encTkt.EndTime.IsZero() && now.After(encTkt.EndTime.UTC().Add(skew)) {
		return nil, nil, fmt.Errorf("gssapi: ticket expired (endtime %s, now %s)", encTkt.EndTime.UTC(), now)
	}
	if !encTkt.StartTime.IsZero() && encTkt.StartTime.UTC().After(now.Add(skew)) {
		return nil, nil, fmt.Errorf("gssapi: ticket not yet valid (starttime %s, now %s)", encTkt.StartTime.UTC(), now)
	}
	if len(encTkt.CAddr) > 0 {
		if len(opts.ClientAddress) == 0 || !ticketAddressMatches(encTkt.CAddr, opts.ClientAddress) {
			return nil, nil, fmt.Errorf("gssapi: client address %s is not authorized by ticket", opts.ClientAddress)
		}
	}
	if encTkt.CRealm != apReq.Ticket.Realm {
		if opts.TransitedPolicy != nil {
			if err := opts.TransitedPolicy(encTkt.CRealm, apReq.Ticket.Realm, encTkt.Transited); err != nil {
				return nil, nil, fmt.Errorf("gssapi: transited-realm policy rejected ticket: %w", err)
			}
		} else if encTkt.Flags.At(messages.TicketFlagTransitCheck) == 0 {
			return nil, nil, fmt.Errorf("gssapi: unchecked cross-realm ticket from %s to %s", encTkt.CRealm, apReq.Ticket.Realm)
		}
	}

	// Decrypt the authenticator with the ticket session key (key usage 11).
	authPlain, err := kerbcrypto.Decrypt(sessionEType, sessionKey, kerbcrypto.KeyUsageAPReqAuthen, apReq.Authenticator.Cipher)
	if err != nil {
		return nil, nil, fmt.Errorf("gssapi: decrypt authenticator: %w", err)
	}
	var auth messages.Authenticator
	if _, err := auth.Unmarshal(authPlain); err != nil {
		return nil, nil, fmt.Errorf("gssapi: parse authenticator: %w", err)
	}

	// RFC 4120 §3.2.3: the authenticator's client name/realm must match the
	// ticket's, binding the presenter to the principal the KDC named.
	if !principalsEqual(auth.CName, encTkt.CName) || auth.CRealm != encTkt.CRealm {
		return nil, nil, fmt.Errorf("gssapi: authenticator client %s@%s does not match ticket client %s@%s",
			principalString(auth.CName), auth.CRealm, principalString(encTkt.CName), encTkt.CRealm)
	}

	// Clock-skew check against the authenticator timestamp (RFC 4120 §3.2.3).
	delta := now.Sub(auth.CTime.UTC())
	if delta < 0 {
		delta = -delta
	}
	if delta > skew {
		return nil, nil, fmt.Errorf("gssapi: authenticator clock skew too large (%s > %s)", delta, skew)
	}

	// Replay check: reject a repeated (client, ctime, cusec) tuple within the skew
	// window (RFC 4120 §3.2.3).
	rc := opts.ReplayCache
	if rc == nil {
		rc = defaultReplayCache
	}
	tuple := replayCacheTuple(apReq.Ticket.SName, apReq.Ticket.Realm, auth)
	if rc.seenBefore(tuple, auth.CTime.UTC().Add(skew), now) {
		return nil, nil, fmt.Errorf("gssapi: replayed authenticator (client %s@%s, ctime %s)", principalString(auth.CName), auth.CRealm, auth.CTime.UTC())
	}

	// Validate the 0x8003 GSS checksum: parse the context flags, and — when the
	// acceptor supplied channel bindings — verify the Bnd field matches
	// (RFC 4121 §4.1.1).
	if err := validateGSSChecksum(&auth, opts.ChannelBindings); err != nil {
		return nil, nil, err
	}

	ctx = &SecContext{
		SessionKey:    sessionKey,
		SessionEType:  sessionEType,
		isAcceptor:    true,
		clientName:    auth.CName,
		clientRealm:   auth.CRealm,
		authenticator: func() *messages.Authenticator { a := auth; return &a }(),
		pacBytes:      extractWin2KPAC(encTkt.AuthorizationData),
		ctime:         auth.CTime.UTC().Truncate(time.Second),
		cusec:         auth.CUSec,
		recvWindow: seqWindow{
			replayDetect: !opts.DisableReplayDetection,
			sequence:     opts.EnforceSequence,
		},
	}

	// Adopt the per-message base key. An acceptor-minted subkey (returned in the
	// AP-REP) wins and sets the AcceptorSubkey flag; otherwise the authenticator
	// subkey (if present) keys per-message tokens without the flag; otherwise the
	// ticket session key does.
	var apRepSubkey *messages.EncryptionKey
	switch {
	case opts.MintSubkey:
		keyLen := kerbcrypto.KeyLen(sessionEType)
		if keyLen <= 0 {
			return nil, nil, fmt.Errorf("gssapi: cannot mint subkey for etype %d", sessionEType)
		}
		sub := make([]byte, keyLen)
		if _, err := rand.Read(sub); err != nil {
			return nil, nil, err
		}
		ctx.SubKey = sub
		ctx.SubKeyEType = sessionEType
		ctx.acceptorSubkey = true
		apRepSubkey = &messages.EncryptionKey{KeyType: sessionEType, KeyValue: sub}
	case auth.SubKey != nil:
		ctx.SubKey = auth.SubKey.KeyValue
		ctx.SubKeyEType = auth.SubKey.KeyType
	}

	// The initiator's authenticator sequence number seeds this side's receive
	// window origin implicitly (the window seeds itself on the first token). The
	// acceptor's own per-message send sequence starts from the number it announces
	// in the AP-REP, mirroring the initiator seeding its send sequence from the
	// authenticator.
	if !mutualRequired(apReq.APOptions) {
		return nil, ctx, nil
	}

	acceptorSeq, err := randomSeq()
	if err != nil {
		return nil, nil, err
	}
	ctx.acceptorSeq = acceptorSeq
	ctx.sendSeq = uint64(acceptorSeq)

	enc := messages.EncAPRepPart{
		CTime:     auth.CTime.UTC(),
		CUSec:     auth.CUSec,
		SubKey:    apRepSubkey,
		SeqNumber: acceptorSeq,
	}
	encBytes, err := enc.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("gssapi: marshal EncAPRepPart: %w", err)
	}
	cipher, err := kerbcrypto.Encrypt(sessionEType, sessionKey, kerbcrypto.KeyUsageAPRepEncPart, encBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("gssapi: encrypt EncAPRepPart: %w", err)
	}
	apRep := messages.APRep{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeAPRep,
		EncPart: messages.EncryptedData{EType: sessionEType, Cipher: cipher},
	}
	apRepBytes, err := apRep.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("gssapi: marshal AP-REP: %w", err)
	}
	outputToken, err = WrapToken(TokIDAPRep, apRepBytes)
	if err != nil {
		return nil, nil, err
	}
	return outputToken, ctx, nil
}

func ticketAddressMatches(addresses []messages.HostAddress, peer net.IP) bool {
	for _, address := range addresses {
		switch address.AddrType {
		case 2: // IPv4
			if ip := peer.To4(); ip != nil && bytes.Equal(ip, address.Address) {
				return true
			}
		case 24: // IPv6
			if ip := peer.To16(); ip != nil && peer.To4() == nil && bytes.Equal(ip, address.Address) {
				return true
			}
		}
	}
	return false
}

func replayCacheTuple(server messages.PrincipalName, serverRealm string, auth messages.Authenticator) string {
	return fmt.Sprintf("%q@%q|%q@%q|%d|%d",
		server.NameString, serverRealm, auth.CName.NameString, auth.CRealm,
		auth.CTime.UTC().Unix(), auth.CUSec)
}

// decryptTicket recovers the EncTicketPart from a ticket by trying each candidate
// service key whose enctype matches the ticket enc-part at key usage 2. Keys come
// from the keytab entry for the ticket service first, then the explicit key list.
func decryptTicket(tkt *messages.Ticket, opts AcceptOptions, useSessionKey bool) (*messages.EncTicketPart, error) {
	etype := tkt.EncPart.EType
	var candidates []ServiceKey
	if useSessionKey {
		if opts.UserToUserKey == nil {
			return nil, fmt.Errorf("gssapi: AP-REQ requires a user-to-user TGT session key")
		}
		if opts.UserToUserKey.EType != etype {
			return nil, fmt.Errorf("gssapi: user-to-user key enctype %d does not match ticket enctype %d", opts.UserToUserKey.EType, etype)
		}
		candidates = append(candidates, *opts.UserToUserKey)
	} else {
		if opts.Keytab != nil {
			principal := principalString(tkt.SName) + "@" + tkt.Realm
			for _, e := range opts.Keytab.Find(principal, etype, -1) {
				candidates = append(candidates, ServiceKey{EType: int(e.EType), Key: e.Key})
			}
		}
		for _, k := range opts.Keys {
			if k.EType == etype {
				candidates = append(candidates, k)
			}
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("gssapi: no service key for ticket enctype %d", etype)
	}
	var lastErr error
	for _, c := range candidates {
		plain, err := kerbcrypto.Decrypt(c.EType, c.Key, kerbcrypto.KeyUsageKDCRepTicket, tkt.EncPart.Cipher)
		if err != nil {
			lastErr = err
			continue
		}
		var enc messages.EncTicketPart
		if _, err := enc.Unmarshal(plain); err != nil {
			lastErr = err
			continue
		}
		return &enc, nil
	}
	if useSessionKey {
		return nil, fmt.Errorf("gssapi: user-to-user key could not decrypt the ticket (enctype %d): %w", etype, lastErr)
	}
	return nil, fmt.Errorf("gssapi: no service key could decrypt the ticket (enctype %d): %w", etype, lastErr)
}

// validateGSSChecksum verifies the authenticator carries a 0x8003 GSS checksum
// and, when the acceptor supplied channel bindings, that the checksum's Bnd field
// matches MD5(channelBindings) (RFC 4121 §4.1.1).
func validateGSSChecksum(auth *messages.Authenticator, channelBindings []byte) error {
	if auth.Cksum == nil {
		return fmt.Errorf("gssapi: AP-REQ authenticator has no checksum")
	}
	if auth.Cksum.CKSumType != ChecksumTypeGSSAPI {
		return fmt.Errorf("gssapi: authenticator checksum type %d is not the GSS 0x8003 type", auth.Cksum.CKSumType)
	}
	if len(auth.Cksum.Checksum) < gss8003MinLen {
		return fmt.Errorf("gssapi: 0x8003 checksum too short (%d bytes)", len(auth.Cksum.Checksum))
	}
	if len(channelBindings) > 0 {
		want := md5.Sum(channelBindings)
		if !hmac.Equal(auth.Cksum.Checksum[4:20], want[:]) {
			return fmt.Errorf("gssapi: channel-binding mismatch (GSS_C_BAD_BINDINGS)")
		}
	}
	return nil
}

// mutualRequired reports whether the AP-options set mutual-required (bit 2, i.e.
// byte 0 / 0x20).
func mutualRequired(apOptions asn1.BitString) bool {
	return len(apOptions.Bytes) > 0 && apOptions.Bytes[0]&0x20 != 0
}

// randomSeq returns a random non-negative 31-bit sequence number, matching the
// initiator's authenticator sequence generation.
func randomSeq() (int, error) {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(buf[:]) & 0x7fffffff), nil
}

// principalsEqual reports whether two principal names have the same name string
// components (the name-type is not compared, matching MIT's krb5_principal_compare
// default which treats an enterprise/principal name-type difference as equal when
// the components match).
func principalsEqual(a, b messages.PrincipalName) bool {
	if len(a.NameString) != len(b.NameString) {
		return false
	}
	for i := range a.NameString {
		if a.NameString[i] != b.NameString[i] {
			return false
		}
	}
	return true
}

// principalString renders a principal name as "comp1/comp2" for diagnostics.
func principalString(p messages.PrincipalName) string {
	var b bytes.Buffer
	for i, c := range p.NameString {
		if i > 0 {
			b.WriteByte('/')
		}
		b.WriteString(c)
	}
	return b.String()
}

// extractWin2KPAC walks the ticket authorization-data for the AD-IF-RELEVANT
// wrapped AD-WIN2K-PAC element and returns the raw PAC bytes, or nil when absent.
func extractWin2KPAC(ad []messages.AuthorizationData) []byte {
	for _, e := range ad {
		if e.ADType != adTypeIfRelevant {
			continue
		}
		var inner []messages.AuthorizationData
		if _, err := asn1.Unmarshal(e.ADData, &inner); err != nil {
			continue
		}
		for _, ie := range inner {
			if ie.ADType == adTypeWin2KPAC {
				return ie.ADData
			}
		}
	}
	return nil
}

// ClientPrincipal returns the authenticated client's principal name and realm,
// taken from the validated AP-REQ authenticator (acceptor side). It is empty on
// an initiator context.
func (ctx *SecContext) ClientPrincipal() (messages.PrincipalName, string) {
	return ctx.clientName, ctx.clientRealm
}

// Authenticator returns the decrypted, validated AP-REQ authenticator (acceptor
// side), or nil on an initiator context. It lets a caller inspect the 0x8003
// checksum, e.g. via ExtractDelegatedCred for an unconstrained-delegation
// KRB-CRED.
func (ctx *SecContext) Authenticator() *messages.Authenticator { return ctx.authenticator }

// HasPAC reports whether the decrypted ticket carried a PAC (acceptor side).
func (ctx *SecContext) HasPAC() bool { return len(ctx.pacBytes) > 0 }

// PACBytes returns the raw AD-WIN2K-PAC bytes extracted from the decrypted ticket
// (acceptor side), or nil when the ticket carried no PAC.
func (ctx *SecContext) PACBytes() []byte { return ctx.pacBytes }

// PAC parses and returns the PAC extracted from the decrypted ticket (acceptor
// side). It returns (nil, nil) when the ticket carried no PAC.
func (ctx *SecContext) PAC() (*pac.PAC, error) {
	if len(ctx.pacBytes) == 0 {
		return nil, nil
	}
	return pac.Parse(ctx.pacBytes)
}
