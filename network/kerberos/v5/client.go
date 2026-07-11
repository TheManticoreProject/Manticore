package kerberos

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credentials"
	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pkinit"
)

// KerberosClient manages Kerberos authentication against an Active Directory KDC.
//
// It provides protocol-level primitives: TGT acquisition with PA-ENC-TIMESTAMP
// pre-authentication, TGS requests, and ASREPRoast. All cryptographic operations
// use the native Manticore implementations (no external Kerberos library).
//
// Typical usage:
//
//	c := kerberos.NewClient("john", "CORP.LOCAL", "10.0.0.1")
//	c.WithPassword("secret")
//	if err := c.GetTGT(); err != nil { ... }
//	ticket, ticketRaw, sessionKey, sessionEType, err := c.GetTGS("cifs/dc01.corp.local", true)
type KerberosClient struct {
	username string
	realm    string
	kdcHost  string

	cred *credentials.Credential

	// Populated after a successful GetTGT call.
	tgtTicket    messages.Ticket
	tgtTicketRaw []byte // raw APPLICATION[1] bytes as received from KDC
	sessionKey   []byte
	sessionEType int
	tgtEnc       messages.EncASRepPart // decrypted AS-REP enc-part: times, flags, sname
	hasTGT       bool

	// stETypes overrides the etype list requested in the TGS-REQ (the service
	// ticket's session-key etype). nil requests AES256, AES128, then RC4.
	stETypes []int

	// preloadedTGS holds service tickets supplied out-of-band (a forged silver
	// ticket, or a captured service ticket) keyed by normalized SPN. When GetTGS
	// is asked for one of these SPNs it returns the preloaded ticket instead of
	// contacting the KDC — enabling silver-ticket use with no TGT.
	preloadedTGS map[string]preloadedServiceTicket

	// serviceTickets caches service tickets the client has obtained (via GetTGS)
	// or been given (LoadServiceTicket), keyed by normalized SPN, together with
	// the enc-part metadata (session key, flags, times, sname/srealm) needed to
	// re-serialize them. It is the source ExportServiceTicketKirbi/CCache read
	// from — the harvest→save half of pass-the-ticket (see export.go).
	serviceTickets map[string]cachedServiceTicket

	// realmKDCs maps an (uppercased) realm to the KDC host to contact when the
	// cross-realm referral chase reaches that realm. Populated via WithRealmKDC.
	realmKDCs map[string]string

	// kdcResolver, if set, resolves an (uppercased) realm to a KDC host for the
	// referral chase. It overrides the default DNS-SRV based resolution and is
	// consulted after realmKDCs and the client's own home realm.
	kdcResolver func(realm string) (string, error)

	// resolver is the DNS resolver used for KDC discovery (SRV lookups) and for
	// resolving KDC hostnames to A/AAAA addresses. nil uses net.DefaultResolver.
	// A custom resolver (see WithResolver) lets callers direct Kerberos DNS at a
	// specific server without touching the host resolver configuration.
	resolver *net.Resolver

	// clockOffset is added to the local clock when stamping outgoing requests. It
	// starts at zero and is set from a KRB_AP_ERR_SKEW reply's server time so a
	// subsequent retry falls inside the KDC's allowable skew window (RFC 4120
	// Section 5.9.1). See now / applyClockSkew.
	clockOffset time.Duration

	// fast, when non-nil, holds an armor TGT enabling RFC 6113 FAST (Kerberos
	// armoring) on the AS exchange. Configured via WithFASTArmor; consumed by
	// GetTGT (see fast.go).
	fast *fastArmor

	// PKINIT (RFC 4556) certificate pre-authentication state, configured via
	// WithPKINIT and consumed by GetTGT (see client_pkinit.go). pkinitReplyKey is
	// the AS reply key derived from the DH exchange, retained for UnPAC-the-hash.
	pkinitPriv       *rsa.PrivateKey
	pkinitCert       []byte
	pkinitGroups     []pkinit.DHGroup
	pkinitReplyKey   []byte
	pkinitReplyEType int
}

// preloadedServiceTicket is a service ticket the client will hand back from
// GetTGS without a KDC round-trip.
type preloadedServiceTicket struct {
	ticket       messages.Ticket
	ticketRaw    []byte
	sessionKey   []byte
	sessionEType int
}

// cachedServiceTicket is a service ticket retained for export: the raw
// APPLICATION[1] ticket bytes plus the KrbCredInfo (session key, flags, times,
// sname/srealm) describing it, ready to feed the .kirbi / ccache marshalers.
type cachedServiceTicket struct {
	ticketRaw []byte
	credInfo  messages.KrbCredInfo
}

// PreferRC4ServiceTicket makes GetTGS request an RC4-HMAC service ticket (RC4
// session key). Windows RPC's DCE-style Kerberos per-message protection is only
// interoperable with RC4 on some servers, so the DCE/RPC client forces it.
func (c *KerberosClient) PreferRC4ServiceTicket() *KerberosClient {
	c.stETypes = []int{messages.ETypeRC4HMAC}
	return c
}

// serviceTicketETypes returns the TGS-REQ etype list: the override if set,
// otherwise the default strongest-first list.
func (c *KerberosClient) serviceTicketETypes() []int {
	if len(c.stETypes) > 0 {
		return c.stETypes
	}
	return []int{
		messages.ETypeAES256CTSHMACSHA196,
		messages.ETypeAES128CTSHMACSHA196,
		messages.ETypeRC4HMAC,
	}
}

// NewClient creates a new KerberosClient for the given username, realm and KDC host.
// The realm is uppercased automatically (required by the Kerberos specification).
// Call WithPassword before calling GetTGT.
func NewClient(username, realm, kdcHost string) *KerberosClient {
	return &KerberosClient{
		username: username,
		realm:    strings.ToUpper(realm),
		kdcHost:  kdcHost,
	}
}

// WithPassword configures a password credential for GetTGT.
// Returns the client to allow fluent chaining.
func (c *KerberosClient) WithPassword(password string) *KerberosClient {
	c.cred = credentials.NewWithPassword(c.username, c.realm, password)
	return c
}

// WithCredential configures an arbitrary credential (password, NT hash, or AES
// key). Returns the client to allow fluent chaining.
func (c *KerberosClient) WithCredential(cred *credentials.Credential) *KerberosClient {
	c.cred = cred
	return c
}

// WithNTHash configures an NT-hash (overpass-the-hash) credential from a hex
// string (accepts an "LM:NT" pair). GetTGT will request an RC4-HMAC TGT.
func (c *KerberosClient) WithNTHash(hexHash string) error {
	cred, err := credentials.NewWithHexNTHash(c.username, c.realm, hexHash)
	if err != nil {
		return err
	}
	c.cred = cred
	return nil
}

// WithAESKey configures a pass-the-key credential from a hex-encoded AES key
// (16 bytes -> AES128, 32 bytes -> AES256).
func (c *KerberosClient) WithAESKey(hexKey string) error {
	cred, err := credentials.NewWithHexAESKey(c.username, c.realm, hexKey)
	if err != nil {
		return err
	}
	c.cred = cred
	return nil
}

// WithResolver installs the DNS resolver used for KDC discovery (SRV lookups)
// and for resolving KDC hostnames to A/AAAA addresses. Passing nil restores the
// default (net.DefaultResolver). This lets a caller point Kerberos DNS at a
// specific server (e.g. the domain controller) without altering system-wide
// resolver configuration. Returns the client to allow fluent chaining.
func (c *KerberosClient) WithResolver(r *net.Resolver) *KerberosClient {
	c.resolver = r
	return c
}

// now returns the current UTC time adjusted by the negotiated clock offset. All
// outgoing timestamps (PA-ENC-TIMESTAMP, Authenticator ctime) are stamped with
// it so a KRB_AP_ERR_SKEW retry lands inside the KDC's skew window.
func (c *KerberosClient) now() time.Time {
	return time.Now().UTC().Add(c.clockOffset)
}

// applyClockSkew records the offset between the client's clock and the KDC's,
// derived from the server time in a KRB_AP_ERR_SKEW reply, so a retry uses a
// timestamp the KDC will accept (RFC 4120 Section 5.9.1). It reports false when
// the error carries no usable server time, in which case no retry is warranted.
func (c *KerberosClient) applyClockSkew(krbErr messages.KRBError) bool {
	if krbErr.STime.IsZero() {
		return false
	}
	// Offset the local clock forward/back to match the KDC. Reference the raw
	// system clock (not c.now()) so repeated applications don't compound.
	c.clockOffset = krbErr.STime.UTC().Sub(time.Now().UTC())
	return true
}

// GetTGT requests a Ticket Granting Ticket from the KDC using the password
// configured via WithPassword.
//
// Windows KDCs silently drop AS-REQs without PA-ENC-TIMESTAMP, so we skip
// the probe and send PA-ENC-TIMESTAMP immediately with the default AD salt
// (realm+username). If the KDC returns PREAUTH_REQUIRED with different
// etype/salt info, we retry once with the corrected values.
func (c *KerberosClient) GetTGT() error {
	// PKINIT (certificate / Shadow Credentials) pre-authentication takes its own
	// path: a DH-based AS exchange with no long-term credential.
	if c.pkinitConfigured() {
		return c.getTGTPKINIT()
	}

	if c.cred == nil {
		return fmt.Errorf("kerberos: no credentials configured: call WithPassword/WithNTHash/WithAESKey/WithCredential first")
	}

	// When a FAST armor TGT is configured, perform an armored AS-REQ with an
	// encrypted-challenge factor (RFC 6113) instead of the bare PA-ENC-TIMESTAMP.
	if c.fast != nil {
		return c.getTGTFAST()
	}

	// Start with the strongest etype the credential can satisfy and the AD
	// default salt. The KDC corrects both via PREAUTH_REQUIRED if needed.
	etype := c.cred.SupportedETypes()[0]
	salt := c.cred.DefaultSalt()

	// At most two attempts: the second only ever follows a KRB_AP_ERR_SKEW, after
	// the clock offset has been applied (RFC 4120 Section 5.9.1).
	for attempt := 0; attempt < 2; attempt++ {
		resp, nonce, err := c.sendASReqWithPreauth(etype, salt, nil)
		if err != nil {
			return err
		}

		var krb_err messages.KRBError
		if _, parse_err := krb_err.Unmarshal(resp); parse_err == nil {
			switch krb_err.ErrorCode {
			case messages.ErrPreauthRequired:
				// The KDC wants a different etype/salt; retry with the corrected
				// values (that path applies its own skew retry).
				etype, salt, s2k_params := c.pickETypeFromError(krb_err)
				return c.doASReqWithPreauth(etype, salt, s2k_params)
			case messages.ErrSkew:
				if attempt == 0 && c.applyClockSkew(krb_err) {
					continue // retry with the KDC-aligned clock
				}
				return fmt.Errorf("kerberos: KDC clock skew too great (error %d): %s", krb_err.ErrorCode, krb_err.EText)
			default:
				return fmt.Errorf("kerberos: KDC error %d: %s", krb_err.ErrorCode, krb_err.EText)
			}
		}

		return c.processASRep(resp, etype, salt, nil, nonce)
	}
	return fmt.Errorf("kerberos: GetTGT: clock-skew retry exhausted")
}

// pacRequestPA returns a PA-PAC-REQUEST PAData element with include-pac = TRUE.
// Windows KDCs require at least this in every AS-REQ to produce a response.
func pacRequestPA() messages.PAData {
	// PA-PAC-REQUEST ::= SEQUENCE { include-pac [0] BOOLEAN }
	// Encoded: 30 05 a0 03 01 01 ff
	return messages.PAData{
		PADataType:  messages.PAPACRequest,
		PADataValue: []byte{0x30, 0x05, 0xa0, 0x03, 0x01, 0x01, 0xff},
	}
}

// GetTGS requests a service ticket for the given Service Principal Name.
// GetTGT must have been called successfully beforehand.
//
// The SPN format is "service/host" (e.g. "cifs/dc01.corp.local") or
// "service/host@REALM".
//
// includePAC controls whether the KDC should include the PAC in the service ticket.
// Pass false for kerberoasting (produces shorter, hashcat-crackable ciphers).
//
// Returns the parsed service Ticket, the raw APPLICATION[1] ticket bytes as
// received from the KDC (suitable for verbatim re-emission in a downstream
// AP-REQ via messages.APReq{TicketRaw: ...}.Marshal), the associated session
// key bytes, and the session key's encryption type. The session-key etype is
// the KDC's choice and is not necessarily the ticket's own encryption type
// (the server long-term key etype): a Windows KDC can, for example, wrap an
// AES256 ticket around an RC4 session key. Callers keying the AP-REQ
// authenticator or per-message GSS tokens must use this returned etype, not
// Ticket.EncPart.EType.
func (c *KerberosClient) GetTGS(spn string, includePAC bool) (messages.Ticket, []byte, []byte, int, error) {
	// A preloaded (forged silver / captured) service ticket for this SPN is
	// returned directly, with no TGT and no KDC round-trip.
	if pt, ok := c.preloadedTGS[normalizeSPN(spn)]; ok {
		return pt.ticket, pt.ticketRaw, pt.sessionKey, pt.sessionEType, nil
	}
	if !c.hasTGT {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: no TGT: call GetTGT first")
	}

	sname, err := parseSPN(spn, c.realm)
	if err != nil {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: parse SPN %q: %w", spn, err)
	}

	// Request the service ticket, chasing any cross-realm referrals returned by
	// the KDC (see chaseServiceTicket / referral.go).
	return c.chaseServiceTicket(sname, includePAC)
}

// Destroy zeroes out key material held by the client.
func (c *KerberosClient) Destroy() {
	for i := range c.sessionKey {
		c.sessionKey[i] = 0
	}
	c.sessionKey = nil
	if c.cred != nil {
		c.cred.Destroy()
	}
	c.hasTGT = false
}

// Username returns the username configured for this client.
func (c *KerberosClient) Username() string { return c.username }

// Realm returns the realm (uppercased) configured for this client.
func (c *KerberosClient) Realm() string { return c.realm }

// KDCHost returns the KDC host configured for this client.
func (c *KerberosClient) KDCHost() string { return c.kdcHost }

// HasTGT reports whether a Ticket Granting Ticket has been acquired (via GetTGT).
func (c *KerberosClient) HasTGT() bool { return c.hasTGT }

// ── internal helpers ──────────────────────────────────────────────────────────

// sendASReq builds and sends an AS-REQ with the given optional PA-DATA slice.
// Returns the raw KDC response bytes and the nonce that was placed in the
// request (the caller must compare it to the nonce in the decrypted
// EncASRepPart per RFC 4120 §3.1.3).
func (c *KerberosClient) sendASReq(pa_data []messages.PAData) ([]byte, int, error) {
	nonce := randomNonce()
	req := &messages.ASReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeASReq,
		PAData:  pa_data,
		ReqBody: messages.KDCReqBody{
			KDCOptions: kdcOptionsForASReq(),
			CName: messages.PrincipalName{
				NameType:   messages.NameTypePrincipal,
				NameString: []string{c.username},
			},
			Realm: c.realm,
			SName: messages.PrincipalName{
				NameType:   messages.NameTypeSRVInst,
				NameString: []string{"krbtgt", c.realm},
			},
			Till:  c.now().Add(24 * time.Hour),
			Nonce: nonce,
			EType: c.cred.SupportedETypes(),
		},
	}

	req_bytes, err := req.Marshal()
	if err != nil {
		return nil, 0, fmt.Errorf("kerberos: marshal AS-REQ: %w", err)
	}
	resp, err := c.sendToRealm(c.realm, req_bytes)
	if err != nil {
		return nil, 0, err
	}
	return resp, nonce, nil
}

// sendToRealm discovers the KDC endpoints serving realm (explicit config, custom
// resolver, or DNS SRV) and sends msg, failing over across them (see
// endpointsForRealm and kdcSendEndpoints).
func (c *KerberosClient) sendToRealm(realm string, msg []byte) ([]byte, error) {
	endpoints, err := c.endpointsForRealm(realm)
	if err != nil {
		return nil, err
	}
	return kdcSendEndpoints(c.resolver, endpoints, msg)
}

// pickETypeFromError extracts the preferred etype, salt and S2KParams from the
// PA-ETYPE-INFO2 structure embedded in a KRBError's EData.
// Falls back to AES-256 with the default AD salt if no EData is present.
func (c *KerberosClient) pickETypeFromError(krb_err messages.KRBError) (int, string, []byte) {
	preferred := c.cred.SupportedETypes()
	default_salt := c.cred.DefaultSalt()
	default_etype := preferred[0]

	if len(krb_err.EData) == 0 {
		return default_etype, default_salt, nil
	}

	// EData may be a SEQUENCE OF PA-DATA or raw ETYPE-INFO2.
	// Try to parse as SEQUENCE OF PA-DATA first.
	var pa_list []messages.PAData
	if _, err := asn1.Unmarshal(krb_err.EData, &pa_list); err == nil {
		for _, pa := range pa_list {
			if pa.PADataType == messages.PAETypeInfo2 {
				var info messages.ETypeInfo2
				if _, err := info.Unmarshal(pa.PADataValue); err == nil && len(info) > 0 {
					return pickBestEType(info, preferred, default_salt)
				}
			}
		}
	}

	// Try to parse EData directly as ETYPE-INFO2.
	var info messages.ETypeInfo2
	if _, err := info.Unmarshal(krb_err.EData); err == nil && len(info) > 0 {
		return pickBestEType(info, preferred, default_salt)
	}

	return default_etype, default_salt, nil
}

// pickBestEType selects, from an ETYPE-INFO2 list, the strongest etype the
// credential can actually satisfy (preferred is the credential's supported set
// in strength order). This prevents choosing, say, AES when only an NT hash is
// held.
func pickBestEType(info messages.ETypeInfo2, preferred []int, default_salt string) (int, string, []byte) {
	for _, want := range preferred {
		for _, entry := range info {
			if entry.EType == want {
				salt := entry.Salt
				if salt == "" {
					salt = default_salt
				}
				return entry.EType, salt, entry.S2KParams
			}
		}
	}
	// No advertised etype is supported by the credential; fall back to the
	// credential's strongest etype with the default salt.
	return preferred[0], default_salt, nil
}

// sendASReqWithPreauth builds and sends an AS-REQ with PA-ENC-TIMESTAMP.
// Returns the raw KDC response bytes and the nonce placed in the request.
func (c *KerberosClient) sendASReqWithPreauth(etype int, salt string, s2k_params []byte) ([]byte, int, error) {
	key, err := c.cred.Key(etype, salt, s2k_params)
	if err != nil {
		return nil, 0, fmt.Errorf("kerberos: derive key: %w", err)
	}

	now := c.now()
	ts := &messages.PAEncTSEnc{
		PATimestamp: now,
		PAUSec:      now.Nanosecond() / 1000,
	}
	ts_bytes, err := ts.Marshal()
	if err != nil {
		return nil, 0, fmt.Errorf("kerberos: marshal PA-ENC-TIMESTAMP: %w", err)
	}

	enc_ts, err := kerbcrypto.Encrypt(etype, key, kerbcrypto.KeyUsageASReqPAEncTimestamp, ts_bytes)
	if err != nil {
		return nil, 0, fmt.Errorf("kerberos: encrypt PA-ENC-TIMESTAMP: %w", err)
	}

	pa_enc_ts := messages.EncryptedData{EType: etype, Cipher: enc_ts}
	pa_enc_ts_bytes, err := asn1.Marshal(pa_enc_ts)
	if err != nil {
		return nil, 0, fmt.Errorf("kerberos: marshal EncryptedData for PA-ENC-TIMESTAMP: %w", err)
	}

	pa_data := []messages.PAData{
		pacRequestPA(),
		{PADataType: messages.PAEncTimestamp, PADataValue: pa_enc_ts_bytes},
	}
	return c.sendASReq(pa_data)
}

// doASReqWithPreauth sends an AS-REQ with PA-ENC-TIMESTAMP and processes the
// AS-REP. On KRB_AP_ERR_SKEW it applies the KDC clock offset and retries once.
func (c *KerberosClient) doASReqWithPreauth(etype int, salt string, s2k_params []byte) error {
	for attempt := 0; attempt < 2; attempt++ {
		resp, nonce, err := c.sendASReqWithPreauth(etype, salt, s2k_params)
		if err != nil {
			return err
		}

		var krb_err messages.KRBError
		if _, parse_err := krb_err.Unmarshal(resp); parse_err == nil {
			if krb_err.ErrorCode == messages.ErrSkew && attempt == 0 && c.applyClockSkew(krb_err) {
				continue // retry with the KDC-aligned clock
			}
			return fmt.Errorf("kerberos: GetTGT failed (error %d): %s", krb_err.ErrorCode, krb_err.EText)
		}

		return c.processASRep(resp, etype, salt, s2k_params, nonce)
	}
	return fmt.Errorf("kerberos: GetTGT: clock-skew retry exhausted")
}

// processASRep decrypts the AS-REP enc-part, verifies that the returned nonce
// matches the one sent in the AS-REQ (RFC 4120 §3.1.3), and stores the TGT
// session key on the client.
func (c *KerberosClient) processASRep(resp []byte, etype int, salt string, s2k_params []byte, request_nonce int) error {
	var as_rep messages.ASRep
	if _, err := as_rep.Unmarshal(resp); err != nil {
		return fmt.Errorf("kerberos: parse AS-REP: %w", err)
	}

	key, err := c.cred.Key(etype, salt, s2k_params)
	if err != nil {
		return fmt.Errorf("kerberos: derive key for AS-REP decrypt: %w", err)
	}

	enc_plain, err := kerbcrypto.Decrypt(etype, key, kerbcrypto.KeyUsageASRepEncPart, as_rep.EncPart.Cipher)
	if err != nil {
		return fmt.Errorf("kerberos: decrypt AS-REP enc-part: %w", err)
	}

	var enc_as_rep messages.EncASRepPart
	if _, err := enc_as_rep.Unmarshal(enc_plain); err != nil {
		return fmt.Errorf("kerberos: parse EncASRepPart: %w", err)
	}

	// RFC 4120 §3.1.3: the nonce in the reply must match the nonce in the
	// request. Rejecting a mismatch defends against replays of captured
	// AS-REPs that happen to decrypt under the current client key.
	if enc_as_rep.Nonce != request_nonce {
		return fmt.Errorf("kerberos: AS-REP nonce mismatch: got %d, want %d", enc_as_rep.Nonce, request_nonce)
	}

	c.tgtTicket = as_rep.Ticket
	c.tgtTicketRaw = as_rep.TicketRaw
	c.sessionKey = enc_as_rep.Key.KeyValue
	c.sessionEType = enc_as_rep.Key.KeyType
	c.tgtEnc = enc_as_rep
	c.hasTGT = true
	return nil
}

// buildAPReq constructs an AP-REQ wrapping the client's current TGT for use in
// TGS-REQ PA-DATA.
func (c *KerberosClient) buildAPReq() ([]byte, error) {
	return c.buildAPReqWith(c.tgtTicket, c.tgtTicketRaw, c.sessionKey, c.sessionEType)
}

// buildAPReqWith constructs an AP-REQ wrapping the given ticket-granting ticket
// and session key. Cross-realm referral chasing presents a different (referral)
// TGT at each hop, so the ticket/key are passed explicitly rather than always
// taken from the client's home TGT. The authenticator's client name and realm
// always identify the original client (they must match the crealm embedded in
// the ticket, which stays the home realm across cross-realm referrals).
func (c *KerberosClient) buildAPReqWith(tgt messages.Ticket, tgtRaw, sessionKey []byte, sessionEType int) ([]byte, error) {
	now := c.now()
	cusec := now.Nanosecond() / 1000

	var seq_buf [4]byte
	if _, err := rand.Read(seq_buf[:]); err != nil {
		return nil, err
	}
	seq_num := int(binary.BigEndian.Uint32(seq_buf[:]) & 0x7fffffff)

	auth := &messages.Authenticator{
		AVno:      messages.KerberosV5,
		CRealm:    c.realm,
		CName:     messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{c.username}},
		CUSec:     cusec,
		CTime:     now,
		SeqNumber: seq_num,
	}

	auth_bytes, err := auth.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal Authenticator: %w", err)
	}
	enc_auth, err := kerbcrypto.Encrypt(sessionEType, sessionKey, kerbcrypto.KeyUsageTGSReqPAAPReqAuthen, auth_bytes)
	if err != nil {
		return nil, fmt.Errorf("encrypt Authenticator: %w", err)
	}

	ap_req := &messages.APReq{
		PVNO:      messages.KerberosV5,
		MsgType:   messages.MsgTypeAPReq,
		APOptions: asn1.BitString{Bytes: []byte{0x00, 0x00, 0x00, 0x00}, BitLength: 32},
		Ticket:    tgt,
		TicketRaw: tgtRaw,
		Authenticator: messages.EncryptedData{
			EType:  sessionEType,
			Cipher: enc_auth,
		},
	}

	return ap_req.Marshal()
}

// Bit positions for KDCOptions (RFC 4120 Section 5.4.1, RFC 6806 for canonicalize).
// Bit 0 is the MSB; bit N sits in byte N/8 at position 7-(N%8).
const (
	kdcOptionForwardable    = 1  // byte 0, 0x40
	kdcOptionProxiable      = 3  // byte 0, 0x10
	kdcOptionRenewable      = 8  // byte 1, 0x80
	kdcOptionCanonicalize   = 15 // byte 1, 0x01 (RFC 6806)
	kdcOptionCNameInAddlTkt = 14 // byte 1, 0x02 (MS-SFU S4U2Proxy)
	kdcOptionRenewableOK    = 27 // byte 3, 0x10
	kdcOptionEncTktInSKey   = 28 // byte 3, 0x08 (user-to-user)
	kdcOptionRenew          = 30 // byte 3, 0x02 (RFC 4120 §3.3.3 renewal)
	kdcOptionValidate       = 31 // byte 3, 0x01 (RFC 4120 §3.3.3 validation)
)

// encodeKDCOptions packs a list of bit positions into a 32-bit BitString
// suitable for use as the KDCOptions field of an AS-REQ or TGS-REQ.
func encodeKDCOptions(bits ...int) asn1.BitString {
	b := make([]byte, 4)
	for _, pos := range bits {
		b[pos/8] |= 1 << (7 - (pos % 8))
	}
	return asn1.BitString{Bytes: b, BitLength: 32}
}

// kdcOptionsForASReq returns the KDCOptions BitString for an AS-REQ, matching
// the flags a real Active Directory client sends (forwardable + proxiable +
// renewable).
func kdcOptionsForASReq() asn1.BitString {
	return encodeKDCOptions(kdcOptionForwardable, kdcOptionProxiable, kdcOptionRenewable)
}

// kdcOptionsForTGSReq returns the KDCOptions BitString for a TGS-REQ. Adds
// canonicalize (RFC 6806) so the KDC canonicalizes the requested SPN — AD
// KDCs can otherwise respond with S_PRINCIPAL_UNKNOWN for non-canonical SPN
// forms — and renewable-ok so the issued service ticket carries the same
// renewable window as the TGT it was derived from.
func kdcOptionsForTGSReq() asn1.BitString {
	return encodeKDCOptions(
		kdcOptionForwardable,
		kdcOptionRenewable,
		kdcOptionCanonicalize,
		kdcOptionRenewableOK,
	)
}

// parseSPN splits a service principal name into a PrincipalName.
// Accepts "service/host", "service/host@REALM" or bare "service".
func parseSPN(spn, default_realm string) (messages.PrincipalName, error) {
	// Strip optional @REALM suffix.
	at := strings.IndexByte(spn, '@')
	if at >= 0 {
		spn = spn[:at]
	}

	slash := strings.IndexByte(spn, '/')
	if slash < 0 {
		return messages.PrincipalName{}, fmt.Errorf("expected format service/host, got %q", spn)
	}
	service := spn[:slash]
	host := spn[slash+1:]
	if service == "" || host == "" {
		return messages.PrincipalName{}, fmt.Errorf("malformed SPN %q", spn)
	}
	return messages.PrincipalName{
		NameType:   messages.NameTypeSRVInst,
		NameString: []string{service, host},
	}, nil
}

// randomNonce returns a random non-negative 31-bit nonce.
func randomNonce() int {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// Fall back to a fixed non-zero value if rand fails.
		return 0x12345678
	}
	return int(binary.BigEndian.Uint32(buf[:]) & 0x7fffffff)
}
