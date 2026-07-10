package kerberos

import (
	"crypto/rsa"
	"fmt"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pkinit"
)

// WithPKINIT configures certificate-based (PKINIT, RFC 4556) pre-authentication
// with Diffie-Hellman key agreement. priv is the client's RSA private key and
// certDER is its DER-encoded X.509 certificate (for Shadow Credentials, a
// self-signed certificate whose public key is registered in the target's
// msDS-KeyCredentialLink). A subsequent GetTGT performs the PKINIT AS exchange
// instead of password/hash pre-authentication.
//
// By default the client offers MODP group 14 (2048-bit) and falls back to group
// 2 (1024-bit); use WithPKINITGroups to override.
func (c *KerberosClient) WithPKINIT(priv *rsa.PrivateKey, certDER []byte) *KerberosClient {
	c.pkinitPriv = priv
	c.pkinitCert = certDER
	if len(c.pkinitGroups) == 0 {
		c.pkinitGroups = []pkinit.DHGroup{pkinit.MODPGroup14(), pkinit.MODPGroup2()}
	}
	return c
}

// WithPKINITGroups overrides the ordered list of MODP Diffie-Hellman groups the
// PKINIT AS exchange will try (the first the KDC accepts is used).
func (c *KerberosClient) WithPKINITGroups(groups ...pkinit.DHGroup) *KerberosClient {
	c.pkinitGroups = groups
	return c
}

// PKINITReplyKey returns the AS reply key derived from the PKINIT Diffie-Hellman
// exchange (nil until GetTGT succeeds over PKINIT). It is the key that decrypts
// the AS-REP enc-part and, for UnPAC-the-hash, the PAC_CREDENTIAL_INFO buffer.
func (c *KerberosClient) PKINITReplyKey() (key []byte, etype int) {
	return c.pkinitReplyKey, c.pkinitReplyEType
}

// pkinitConfigured reports whether PKINIT pre-authentication is configured.
func (c *KerberosClient) pkinitConfigured() bool {
	return c.pkinitPriv != nil && len(c.pkinitCert) > 0
}

// getTGTPKINIT performs the PKINIT AS exchange (RFC 4556) to obtain a TGT. It
// tries each configured MODP DH group in order until the KDC issues a reply.
func (c *KerberosClient) getTGTPKINIT() error {
	groups := c.pkinitGroups
	if len(groups) == 0 {
		groups = []pkinit.DHGroup{pkinit.MODPGroup14(), pkinit.MODPGroup2()}
	}
	var lastErr error
	for _, group := range groups {
		err := c.doPKINITASReq(group)
		if err == nil {
			return nil
		}
		lastErr = err
	}
	return fmt.Errorf("kerberos: PKINIT AS exchange failed: %w", lastErr)
}

// doPKINITASReq builds and sends one PKINIT AS-REQ with the given DH group and
// processes the reply.
func (c *KerberosClient) doPKINITASReq(group pkinit.DHGroup) error {
	for attempt := 0; attempt < 2; attempt++ {
		bodyNonce := randomNonce()
		body := messages.KDCReqBody{
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
			Nonce: bodyNonce,
			EType: []int{
				messages.ETypeAES256CTSHMACSHA196,
				messages.ETypeAES128CTSHMACSHA196,
				messages.ETypeRC4HMAC,
			},
		}

		bodyDER, err := messages.EncodeKDCReqBody(body)
		if err != nil {
			return fmt.Errorf("kerberos: encode KDC-REQ-BODY: %w", err)
		}

		// Windows echoes the PKAuthenticator nonce (not the KDC-REQ-BODY nonce)
		// into the AS-REP enc-part, so use the same nonce for both to keep the
		// reply-nonce check meaningful.
		paValue, pkReq, err := pkinit.BuildASReqPAData(bodyDER, c.pkinitPriv, c.pkinitCert, group, bodyNonce, c.now())
		if err != nil {
			return err
		}

		req := &messages.ASReq{
			PVNO:    messages.KerberosV5,
			MsgType: messages.MsgTypeASReq,
			PAData: []messages.PAData{
				{PADataType: messages.PAPKASReq, PADataValue: paValue},
				pacRequestPA(),
			},
			ReqBody: body,
		}
		reqBytes, err := req.Marshal()
		if err != nil {
			return fmt.Errorf("kerberos: marshal PKINIT AS-REQ: %w", err)
		}
		resp, err := c.sendToRealm(c.realm, reqBytes)
		if err != nil {
			return err
		}

		var krbErr messages.KRBError
		if _, parseErr := krbErr.Unmarshal(resp); parseErr == nil {
			if krbErr.ErrorCode == messages.ErrSkew && attempt == 0 && c.applyClockSkew(krbErr) {
				continue // retry with the KDC-aligned clock
			}
			return fmt.Errorf("kerberos: PKINIT KDC error %d: %s", krbErr.ErrorCode, krbErr.EText)
		}

		return c.processPKINITASRep(resp, pkReq, bodyNonce)
	}
	return fmt.Errorf("kerberos: PKINIT: clock-skew retry exhausted")
}

// processPKINITASRep parses the PKINIT AS-REP, derives the DH reply key, decrypts
// the enc-part and stores the resulting TGT on the client.
func (c *KerberosClient) processPKINITASRep(resp []byte, pkReq *pkinit.Request, bodyNonce int) error {
	var asRep messages.ASRep
	if _, err := asRep.Unmarshal(resp); err != nil {
		return fmt.Errorf("kerberos: parse PKINIT AS-REP: %w", err)
	}

	// The KDC returns its DH contribution in a PA-PK-AS-REP PA-DATA element.
	var paValue []byte
	for _, pa := range asRep.PAData {
		if pa.PADataType == messages.PAPKASRep || pa.PADataType == messages.PAPKASRepOld {
			paValue = pa.PADataValue
			break
		}
	}
	if len(paValue) == 0 {
		return fmt.Errorf("kerberos: PKINIT AS-REP has no PA-PK-AS-REP element")
	}

	reply, err := pkinit.ParseASRepPAData(paValue)
	if err != nil {
		return err
	}

	etype := asRep.EncPart.EType
	keyLen := kerbcrypto.KeyLen(etype)
	if keyLen == 0 {
		return fmt.Errorf("kerberos: PKINIT AS-REP unsupported reply-key etype %d", etype)
	}

	// RFC 4556 leaves the nonce mixing in octetstring2key ambiguous for a
	// non-reused exchange; try each candidate reply key until one decrypts.
	candidates, err := pkReq.ReplyKeyCandidates(reply, keyLen)
	if err != nil {
		return err
	}
	var lastErr error
	for _, key := range candidates {
		encPlain, decErr := kerbcrypto.Decrypt(etype, key, kerbcrypto.KeyUsageASRepEncPart, asRep.EncPart.Cipher)
		if decErr != nil {
			lastErr = decErr
			continue
		}
		var encASRep messages.EncASRepPart
		if _, err := encASRep.Unmarshal(encPlain); err != nil {
			lastErr = err
			continue
		}
		if encASRep.Nonce != bodyNonce {
			lastErr = fmt.Errorf("kerberos: PKINIT AS-REP nonce mismatch")
			continue
		}

		c.tgtTicket = asRep.Ticket
		c.tgtTicketRaw = asRep.TicketRaw
		c.sessionKey = encASRep.Key.KeyValue
		c.sessionEType = encASRep.Key.KeyType
		c.tgtEnc = encASRep
		c.hasTGT = true
		c.pkinitReplyKey = key
		c.pkinitReplyEType = etype
		return nil
	}
	return fmt.Errorf("kerberos: PKINIT AS-REP could not be decrypted with the DH-derived reply key: %w", lastErr)
}
