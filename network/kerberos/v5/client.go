package kerberos

import (
	"crypto/rand"
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/messages"
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
//	ticket, sessionKey, err := c.GetTGS("cifs/dc01.corp.local", true)
type KerberosClient struct {
	username string
	realm    string
	kdcHost  string

	password string

	// Populated after a successful GetTGT call.
	tgtTicket    messages.Ticket
	tgtTicketRaw []byte // raw APPLICATION[1] bytes as received from KDC
	sessionKey   []byte
	sessionEType int
	hasTGT       bool
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

// WithPassword stores the password for use in GetTGT.
// Returns the client to allow fluent chaining.
func (c *KerberosClient) WithPassword(password string) *KerberosClient {
	c.password = password
	return c
}

// WithCCache is not yet supported in the native implementation.
// Use the gokrb5-backed KerberosInit helper for ccache-based LDAP binds.
func (c *KerberosClient) WithCCache(_ string) error {
	return fmt.Errorf("kerberos: ccache not supported in native implementation; use gokrb5 KerberosInit for LDAP GSSAPI binds")
}

// GetTGT requests a Ticket Granting Ticket from the KDC using the password
// configured via WithPassword.
//
// Windows KDCs silently drop AS-REQs without PA-ENC-TIMESTAMP, so we skip
// the probe and send PA-ENC-TIMESTAMP immediately with the default AD salt
// (realm+username). If the KDC returns PREAUTH_REQUIRED with different
// etype/salt info, we retry once with the corrected values.
func (c *KerberosClient) GetTGT() error {
	if c.password == "" {
		return fmt.Errorf("kerberos: no credentials configured: call WithPassword first")
	}

	// Attempt with default AD salt. The KDC may respond with PREAUTH_REQUIRED
	// if a different salt or etype is required.
	etype := messages.ETypeAES256CTSHMACSHA196
	salt := c.realm + c.username
	resp, err := c.sendASReqWithPreauth(etype, salt, nil)
	if err != nil {
		return err
	}

	// If the KDC requires a different etype/salt, it responds with PREAUTH_REQUIRED.
	var krb_err messages.KRBError
	if _, parse_err := krb_err.Unmarshal(resp); parse_err == nil {
		if krb_err.ErrorCode == messages.ErrPreauthRequired {
			etype, salt, s2k_params := c.pickETypeFromError(krb_err)
			return c.doASReqWithPreauth(etype, salt, s2k_params)
		}
		return fmt.Errorf("kerberos: KDC error %d: %s", krb_err.ErrorCode, krb_err.EText)
	}

	return c.processASRep(resp, etype, salt, nil)
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
// Returns the service Ticket and its associated session key bytes.
func (c *KerberosClient) GetTGS(spn string, includePAC bool) (messages.Ticket, []byte, error) {
	if !c.hasTGT {
		return messages.Ticket{}, nil, fmt.Errorf("kerberos: no TGT: call GetTGT first")
	}

	sname, err := parseSPN(spn, c.realm)
	if err != nil {
		return messages.Ticket{}, nil, fmt.Errorf("kerberos: parse SPN %q: %w", spn, err)
	}

	// Build AP-REQ wrapping the TGT.
	ap_req_bytes, err := c.buildAPReq()
	if err != nil {
		return messages.Ticket{}, nil, fmt.Errorf("kerberos: build AP-REQ: %w", err)
	}

	// Build TGS-REQ.
	nonce := randomNonce()
	// PA-PAC-REQUEST: SEQUENCE { [0] BOOLEAN } — TRUE=0xff, FALSE=0x00
	pacBool := byte(0xff)
	if !includePAC {
		pacBool = 0x00
	}
	tgs_req := &messages.TGSReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeTGSReq,
		PAData: []messages.PAData{
			{PADataType: messages.PATGSReq, PADataValue: ap_req_bytes},
			{PADataType: messages.PAPACRequest, PADataValue: []byte{0x30, 0x05, 0xa0, 0x03, 0x01, 0x01, pacBool}},
		},
		ReqBody: messages.KDCReqBody{
			KDCOptions: kdcOptionsForwardable(),
			Realm:      c.realm,
			SName:      sname,
			Till:       time.Now().UTC().Add(24 * time.Hour),
			Nonce:      nonce,
			EType: []int{
				messages.ETypeAES256CTSHMACSHA196,
				messages.ETypeAES128CTSHMACSHA196,
				messages.ETypeRC4HMAC,
			},
		},
	}

	tgs_req_bytes, err := tgs_req.Marshal()
	if err != nil {
		return messages.Ticket{}, nil, fmt.Errorf("kerberos: marshal TGS-REQ: %w", err)
	}
	resp, err := kdcSend(c.kdcHost, defaultKDCPort, tgs_req_bytes)
	if err != nil {
		return messages.Ticket{}, nil, err
	}

	// Check for KRBError.
	var krb_err messages.KRBError
	if _, parse_err := krb_err.Unmarshal(resp); parse_err == nil {
		return messages.Ticket{}, nil, fmt.Errorf("kerberos: TGS error %d: %s", krb_err.ErrorCode, krb_err.EText)
	}

	// Parse TGS-REP.
	var tgs_rep messages.TGSRep
	if _, err := tgs_rep.Unmarshal(resp); err != nil {
		return messages.Ticket{}, nil, fmt.Errorf("kerberos: parse TGS-REP: %w", err)
	}

	// Decrypt enc-part with the TGT session key.
	enc_plain, err := kerbcrypto.Decrypt(
		c.sessionEType,
		c.sessionKey,
		kerbcrypto.KeyUsageTGSRepEncSessionKey,
		tgs_rep.EncPart.Cipher,
	)
	if err != nil {
		return messages.Ticket{}, nil, fmt.Errorf("kerberos: decrypt TGS-REP enc-part: %w", err)
	}

	var enc_tgs_rep messages.EncTGSRepPart
	if _, err := enc_tgs_rep.Unmarshal(enc_plain); err != nil {
		return messages.Ticket{}, nil, fmt.Errorf("kerberos: parse EncTGSRepPart: %w", err)
	}

	return tgs_rep.Ticket, enc_tgs_rep.Key.KeyValue, nil
}

// Destroy zeroes out key material held by the client.
func (c *KerberosClient) Destroy() {
	for i := range c.sessionKey {
		c.sessionKey[i] = 0
	}
	c.sessionKey = nil
	c.password = ""
	c.hasTGT = false
}

// Username returns the username configured for this client.
func (c *KerberosClient) Username() string { return c.username }

// Realm returns the realm (uppercased) configured for this client.
func (c *KerberosClient) Realm() string { return c.realm }

// KDCHost returns the KDC host configured for this client.
func (c *KerberosClient) KDCHost() string { return c.kdcHost }

// ── internal helpers ──────────────────────────────────────────────────────────

// sendASReq builds and sends an AS-REQ with the given optional PA-DATA slice.
// Returns the raw KDC response bytes.
func (c *KerberosClient) sendASReq(pa_data []messages.PAData) ([]byte, error) {
	req := &messages.ASReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeASReq,
		PAData:  pa_data,
		ReqBody: messages.KDCReqBody{
			KDCOptions: kdcOptionsForwardable(),
			CName: messages.PrincipalName{
				NameType:   messages.NameTypePrincipal,
				NameString: []string{c.username},
			},
			Realm: c.realm,
			SName: messages.PrincipalName{
				NameType:   messages.NameTypeSRVInst,
				NameString: []string{"krbtgt", c.realm},
			},
			Till:  time.Now().UTC().Add(24 * time.Hour),
			Nonce: randomNonce(),
			EType: []int{
				messages.ETypeAES256CTSHMACSHA196,
				messages.ETypeAES128CTSHMACSHA196,
				messages.ETypeRC4HMAC,
			},
		},
	}

	req_bytes, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal AS-REQ: %w", err)
	}
	return kdcSend(c.kdcHost, defaultKDCPort, req_bytes)
}

// pickETypeFromError extracts the preferred etype, salt and S2KParams from the
// PA-ETYPE-INFO2 structure embedded in a KRBError's EData.
// Falls back to AES-256 with the default AD salt if no EData is present.
func (c *KerberosClient) pickETypeFromError(krb_err messages.KRBError) (int, string, []byte) {
	default_salt := c.realm + c.username
	default_etype := messages.ETypeAES256CTSHMACSHA196

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
					return pickBestEType(info, default_salt)
				}
			}
		}
	}

	// Try to parse EData directly as ETYPE-INFO2.
	var info messages.ETypeInfo2
	if _, err := info.Unmarshal(krb_err.EData); err == nil && len(info) > 0 {
		return pickBestEType(info, default_salt)
	}

	return default_etype, default_salt, nil
}

// pickBestEType selects the strongest supported etype from an ETypeInfo2 list.
func pickBestEType(info messages.ETypeInfo2, default_salt string) (int, string, []byte) {
	// Preference order: AES256 > AES128 > RC4.
	preferred := []int{
		messages.ETypeAES256CTSHMACSHA196,
		messages.ETypeAES128CTSHMACSHA196,
		messages.ETypeRC4HMAC,
	}
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
	// Fallback to first entry.
	e := info[0]
	if e.Salt == "" {
		e.Salt = default_salt
	}
	return e.EType, e.Salt, e.S2KParams
}

// sendASReqWithPreauth builds and sends an AS-REQ with PA-ENC-TIMESTAMP.
// Returns the raw KDC response bytes.
func (c *KerberosClient) sendASReqWithPreauth(etype int, salt string, s2k_params []byte) ([]byte, error) {
	key, err := kerbcrypto.StringToKey(etype, c.password, salt, s2k_params)
	if err != nil {
		return nil, fmt.Errorf("kerberos: StringToKey: %w", err)
	}

	now := time.Now().UTC()
	ts := &messages.PAEncTSEnc{
		PATimestamp: now,
		PAUSec:      now.Nanosecond() / 1000,
	}
	ts_bytes, err := ts.Marshal()
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal PA-ENC-TIMESTAMP: %w", err)
	}

	enc_ts, err := kerbcrypto.Encrypt(etype, key, kerbcrypto.KeyUsageASReqPAEncTimestamp, ts_bytes)
	if err != nil {
		return nil, fmt.Errorf("kerberos: encrypt PA-ENC-TIMESTAMP: %w", err)
	}

	pa_enc_ts := messages.EncryptedData{EType: etype, Cipher: enc_ts}
	pa_enc_ts_bytes, err := asn1.Marshal(pa_enc_ts)
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal EncryptedData for PA-ENC-TIMESTAMP: %w", err)
	}

	pa_data := []messages.PAData{
		pacRequestPA(),
		{PADataType: messages.PAEncTimestamp, PADataValue: pa_enc_ts_bytes},
	}
	return c.sendASReq(pa_data)
}

// doASReqWithPreauth sends an AS-REQ with PA-ENC-TIMESTAMP and processes the AS-REP.
func (c *KerberosClient) doASReqWithPreauth(etype int, salt string, s2k_params []byte) error {
	resp, err := c.sendASReqWithPreauth(etype, salt, s2k_params)
	if err != nil {
		return err
	}

	var krb_err messages.KRBError
	if _, parse_err := krb_err.Unmarshal(resp); parse_err == nil {
		return fmt.Errorf("kerberos: GetTGT failed (error %d): %s", krb_err.ErrorCode, krb_err.EText)
	}

	return c.processASRep(resp, etype, salt, s2k_params)
}

// processASRep decrypts the AS-REP enc-part and stores the TGT session key.
func (c *KerberosClient) processASRep(resp []byte, etype int, salt string, s2k_params []byte) error {
	var as_rep messages.ASRep
	if _, err := as_rep.Unmarshal(resp); err != nil {
		return fmt.Errorf("kerberos: parse AS-REP: %w", err)
	}

	key, err := kerbcrypto.StringToKey(etype, c.password, salt, s2k_params)
	if err != nil {
		return fmt.Errorf("kerberos: StringToKey for AS-REP decrypt: %w", err)
	}

	enc_plain, err := kerbcrypto.Decrypt(etype, key, kerbcrypto.KeyUsageASRepEncPart, as_rep.EncPart.Cipher)
	if err != nil {
		return fmt.Errorf("kerberos: decrypt AS-REP enc-part: %w", err)
	}

	var enc_as_rep messages.EncASRepPart
	if _, err := enc_as_rep.Unmarshal(enc_plain); err != nil {
		return fmt.Errorf("kerberos: parse EncASRepPart: %w", err)
	}

	c.tgtTicket = as_rep.Ticket
	c.tgtTicketRaw = as_rep.TicketRaw
	c.sessionKey = enc_as_rep.Key.KeyValue
	c.sessionEType = enc_as_rep.Key.KeyType
	c.hasTGT = true
	return nil
}

// buildAPReq constructs an AP-REQ wrapping the TGT for use in TGS-REQ PA-DATA.
func (c *KerberosClient) buildAPReq() ([]byte, error) {
	now := time.Now().UTC()
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
	enc_auth, err := kerbcrypto.Encrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageTGSReqPAAPReqAuthen, auth_bytes)
	if err != nil {
		return nil, fmt.Errorf("encrypt Authenticator: %w", err)
	}

	ap_req := &messages.APReq{
		PVNO:      messages.KerberosV5,
		MsgType:   messages.MsgTypeAPReq,
		APOptions: asn1.BitString{Bytes: []byte{0x00, 0x00, 0x00, 0x00}, BitLength: 32},
		Ticket:    c.tgtTicket,
		TicketRaw: c.tgtTicketRaw,
		Authenticator: messages.EncryptedData{
			EType:  c.sessionEType,
			Cipher: enc_auth,
		},
	}

	return ap_req.Marshal()
}

// kdcOptionsForwardable returns a KDCOptions BitString with the forwardable flag set.
// Bit positions follow RFC 4120 Section 5.4.1 (bit 0 = MSB).
func kdcOptionsForwardable() asn1.BitString {
	// Forwardable = bit 1 (RFC 4120), renewable-ok = bit 27.
	// Encoded as a 32-bit big-endian bit string: bit 1 → 0x40 in first byte.
	return asn1.BitString{
		Bytes:     []byte{0x40, 0x00, 0x00, 0x00},
		BitLength: 32,
	}
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
