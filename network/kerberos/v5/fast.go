package kerberos

import (
	"crypto/rand"
	"encoding/asn1"
	"fmt"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// This file implements RFC 6113 (Kerberos FAST — Flexible Authentication Secure
// Tunneling), as profiled by MS-KILE for "Kerberos armoring", on the AS
// exchange. FAST wraps the real AS-REQ inside a PA-FX-FAST envelope keyed by an
// armor key that is derived from an armor TGT, and carries the client's
// credential proof as a PA-ENCRYPTED-CHALLENGE FAST factor rather than the
// bare PA-ENC-TIMESTAMP of RFC 4120.
//
// KRB-FX-CF2 pepper strings (RFC 6113 §5.4.1.1 / §5.4.3 / §5.4.6).
const (
	fastPepperSubkeyArmor      = "subkeyarmor"
	fastPepperTicketArmor      = "ticketarmor"
	fastPepperClientChallenge  = "clientchallengearmor"
	fastPepperKDCChallenge     = "kdcchallengearmor"
	fastPepperChallengeLongTem = "challengelongterm"
	fastPepperStrengthenKey    = "strengthenkey"
	fastPepperReplyKey         = "replykey"
)

// fastArmor holds the armor TGT and its session key used to armor a FAST
// exchange (RFC 6113 §5.4.1.1, FX_FAST_ARMOR_AP_REQUEST).
type fastArmor struct {
	cname        string
	realm        string
	ticket       messages.Ticket
	ticketRaw    []byte
	sessionKey   []byte
	sessionEType int
}

// WithFASTArmor enables RFC 6113 FAST (Kerberos armoring) on the AS exchange,
// using the supplied armor TGT. cname/realm identify the armor principal (the
// account that owns the TGT); ticket/ticketRaw are the armor TGT (raw is the
// verbatim APPLICATION[1] bytes as issued by the KDC); sessionKey/sessionEType
// are the armor TGT's session key. Once configured, GetTGT performs a
// FAST-armored AS-REQ with a PA-ENCRYPTED-CHALLENGE factor. Returns the client
// for fluent chaining.
func (c *KerberosClient) WithFASTArmor(cname, realm string, ticket messages.Ticket, ticketRaw, sessionKey []byte, sessionEType int) *KerberosClient {
	c.fast = &fastArmor{
		cname:        cname,
		realm:        realm,
		ticket:       ticket,
		ticketRaw:    ticketRaw,
		sessionKey:   append([]byte(nil), sessionKey...),
		sessionEType: sessionEType,
	}
	return c
}

// WithFASTArmorFromClient enables FAST using another client's already-acquired
// TGT as the armor (the "self-armor" pattern when armor is the same client that
// has a TGT). The armor client MUST have completed GetTGT. Returns the client
// for fluent chaining.
func (c *KerberosClient) WithFASTArmorFromClient(armor *KerberosClient) *KerberosClient {
	return c.WithFASTArmor(armor.username, armor.realm, armor.tgtTicket, armor.tgtTicketRaw, armor.sessionKey, armor.sessionEType)
}

// FASTEnabled reports whether a FAST armor TGT has been configured.
func (c *KerberosClient) FASTEnabled() bool { return c.fast != nil }

// getTGTFAST performs a FAST-armored AS-REQ with a PA-ENCRYPTED-CHALLENGE factor
// (RFC 6113 §5.4.6) and stores the resulting TGT on success. It is invoked by
// GetTGT when a FAST armor has been configured via WithFASTArmor.
func (c *KerberosClient) getTGTFAST() error {
	if c.cred == nil {
		return fmt.Errorf("kerberos: no credentials configured: call WithPassword/WithNTHash/WithAESKey/WithCredential first")
	}
	armor := c.fast

	// Derive the armor key once: a fresh subkey combined with the armor TGT
	// session key (RFC 6113 §5.4.1.1). The same armor key protects the request
	// envelope and decrypts the reply.
	subkey := make([]byte, kerbcrypto.KeyLen(armor.sessionEType))
	if _, err := rand.Read(subkey); err != nil {
		return fmt.Errorf("kerberos: FAST subkey: %w", err)
	}
	armorKey, armorEType, err := kerbcrypto.KRBFXCF2(
		subkey, armor.sessionEType,
		armor.sessionKey, armor.sessionEType,
		fastPepperSubkeyArmor, fastPepperTicketArmor)
	if err != nil {
		return fmt.Errorf("kerberos: derive FAST armor key: %w", err)
	}

	// Build the FX_FAST_ARMOR_AP_REQUEST once; it carries the subkey the KDC
	// needs to reconstruct the armor key. It is reused across retries.
	apReq, err := c.buildArmorAPReq(subkey)
	if err != nil {
		return fmt.Errorf("kerberos: build FAST armor AP-REQ: %w", err)
	}
	fastArmorStruct := &messages.KrbFastArmor{ArmorType: messages.FXFastArmorAPRequest, ArmorValue: apReq}

	// The client's long-term key starts at the strongest supported etype with
	// the AD default salt; a PREAUTH error corrects both via PA-ETYPE-INFO2.
	clientEType := c.cred.SupportedETypes()[0]
	salt := c.cred.DefaultSalt()
	var s2kParams []byte
	var cookie *messages.PAData

	// At most three attempts: one for a cookie/etype correction, one for a
	// clock-skew correction, plus the initial try.
	for attempt := 0; attempt < 3; attempt++ {
		nonce := randomNonce()
		resp, err := c.sendFASTAS(fastArmorStruct, armorKey, armorEType, clientEType, salt, s2kParams, nonce, cookie)
		if err != nil {
			return err
		}

		var krbErr messages.KRBError
		if _, parseErr := krbErr.Unmarshal(resp); parseErr == nil {
			innerErr, newCookie, info := c.parseFASTError(krbErr, armorKey, armorEType)
			if newCookie != nil {
				cookie = newCookie
			}
			// A PA-ETYPE-INFO2 inside the armored error corrects the client key.
			if info != nil {
				clientEType, salt, s2kParams = pickBestEType(info, c.cred.SupportedETypes(), c.cred.DefaultSalt())
			}
			effective := krbErr
			if innerErr != nil {
				effective = *innerErr
			}
			switch effective.ErrorCode {
			case messages.ErrPreauthRequired:
				if attempt < 2 {
					continue // retry with the corrected etype/salt and cookie
				}
			case messages.ErrSkew:
				if attempt < 2 && c.applyClockSkew(effective) {
					continue
				}
			}
			return fmt.Errorf("kerberos: FAST GetTGT failed (error %d): %s", effective.ErrorCode, effective.EText)
		}

		return c.processFASTASRep(resp, armorKey, armorEType, clientEType, salt, s2kParams, nonce)
	}
	return fmt.Errorf("kerberos: FAST GetTGT: retry budget exhausted")
}

// buildArmorAPReq builds the FX_FAST_ARMOR_AP_REQUEST AP-REQ over the armor TGT.
// The authenticator MUST carry the subkey (RFC 6113 §5.4.1.1); it is encrypted
// with the armor TGT session key under key usage 11 (a standalone AP-REQ, not a
// PA-TGS-REQ).
func (c *KerberosClient) buildArmorAPReq(subkey []byte) ([]byte, error) {
	armor := c.fast
	now := c.now()

	auth := &messages.Authenticator{
		AVno:   messages.KerberosV5,
		CRealm: armor.realm,
		CName:  messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{armor.cname}},
		CUSec:  now.Nanosecond() / 1000,
		CTime:  now,
		SubKey: &messages.EncryptionKey{KeyType: armor.sessionEType, KeyValue: subkey},
	}
	authBytes, err := auth.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal armor Authenticator: %w", err)
	}
	encAuth, err := kerbcrypto.Encrypt(armor.sessionEType, armor.sessionKey, kerbcrypto.KeyUsageAPReqAuthen, authBytes)
	if err != nil {
		return nil, fmt.Errorf("encrypt armor Authenticator: %w", err)
	}

	apReq := &messages.APReq{
		PVNO:      messages.KerberosV5,
		MsgType:   messages.MsgTypeAPReq,
		APOptions: asn1.BitString{Bytes: []byte{0x00, 0x00, 0x00, 0x00}, BitLength: 32},
		Ticket:    armor.ticket,
		TicketRaw: armor.ticketRaw,
		Authenticator: messages.EncryptedData{
			EType:  armor.sessionEType,
			Cipher: encAuth,
		},
	}
	return apReq.Marshal()
}

// asReqBody builds the AS-REQ KDC-REQ-BODY requesting a TGT for this client.
func (c *KerberosClient) asReqBody(nonce int) messages.KDCReqBody {
	return messages.KDCReqBody{
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
	}
}

// sendFASTAS assembles and sends one FAST-armored AS-REQ with a
// PA-ENCRYPTED-CHALLENGE factor, returning the raw KDC response.
func (c *KerberosClient) sendFASTAS(armor *messages.KrbFastArmor, armorKey []byte, armorEType, clientEType int, salt string, s2kParams []byte, nonce int, cookie *messages.PAData) ([]byte, error) {
	// PA-ENCRYPTED-CHALLENGE: PA-ENC-TS-ENC encrypted in the challenge key,
	// which combines the armor key with the client long-term key (RFC 6113
	// §5.4.6).
	clientKey, err := c.cred.Key(clientEType, salt, s2kParams)
	if err != nil {
		return nil, fmt.Errorf("kerberos: derive client key: %w", err)
	}
	challengeKey, challEType, err := kerbcrypto.KRBFXCF2(
		armorKey, armorEType, clientKey, clientEType,
		fastPepperClientChallenge, fastPepperChallengeLongTem)
	if err != nil {
		return nil, fmt.Errorf("kerberos: derive challenge key: %w", err)
	}

	now := c.now()
	tsBytes, err := (&messages.PAEncTSEnc{PATimestamp: now, PAUSec: now.Nanosecond() / 1000}).Marshal()
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal PA-ENC-TS-ENC: %w", err)
	}
	encChallenge, err := kerbcrypto.Encrypt(challEType, challengeKey, kerbcrypto.KeyUsageEncChallengeClient, tsBytes)
	if err != nil {
		return nil, fmt.Errorf("kerberos: encrypt challenge: %w", err)
	}
	challengeValue, err := asn1.Marshal(messages.EncryptedData{EType: challEType, Cipher: encChallenge})
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal EncryptedChallenge: %w", err)
	}

	// Inner (armor-protected) padata and request body — the KDC uses these in
	// preference to the outer ones.
	innerPAData := []messages.PAData{
		pacRequestPA(),
		{PADataType: messages.PAEncryptedChallenge, PADataValue: challengeValue},
	}
	if cookie != nil {
		innerPAData = append(innerPAData, *cookie)
	}
	innerBody := c.asReqBody(nonce)

	fastReq := &messages.KrbFastReq{
		FastOptions: messages.NewKerberosFlags(),
		PAData:      innerPAData,
		ReqBody:     innerBody,
	}
	fastReqBytes, err := fastReq.Marshal()
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal KrbFastReq: %w", err)
	}
	encFastReq, err := kerbcrypto.Encrypt(armorEType, armorKey, kerbcrypto.KeyUsageFASTEnc, fastReqBytes)
	if err != nil {
		return nil, fmt.Errorf("kerberos: encrypt KrbFastReq: %w", err)
	}

	// The req-checksum is a keyed checksum over the OUTER KDC-REQ-BODY, computed
	// with the armor key (RFC 6113 §5.4.2, key usage 50). Encode the outer body
	// once so the transmitted bytes are byte-identical to what was checksummed.
	outerBody := c.asReqBody(randomNonce())
	outerBodyBytes, err := messages.EncodeKDCReqBody(outerBody)
	if err != nil {
		return nil, fmt.Errorf("kerberos: encode outer KDC-REQ-BODY: %w", err)
	}
	cksumType, ok := kerbcrypto.ChecksumTypeForEType(armorEType)
	if !ok {
		return nil, fmt.Errorf("kerberos: no checksum type for armor etype %d", armorEType)
	}
	cksum, err := kerbcrypto.GetChecksum(cksumType, armorKey, kerbcrypto.KeyUsageFASTReqChksum, outerBodyBytes)
	if err != nil {
		return nil, fmt.Errorf("kerberos: FAST req-checksum: %w", err)
	}

	armoredReq := &messages.KrbFastArmoredReq{
		Armor:       armor,
		ReqChecksum: messages.Checksum{CKSumType: cksumType, Checksum: cksum},
		EncFastReq:  messages.EncryptedData{EType: armorEType, Cipher: encFastReq},
	}
	fastValue, err := messages.MarshalPAFXFastRequest(armoredReq)
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal PA-FX-FAST: %w", err)
	}

	// The outer AS-REQ carries only PA-FX-FAST; the KDC ignores the rest.
	asReq := &messages.ASReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeASReq,
		PAData:  []messages.PAData{{PADataType: messages.PAFXFast, PADataValue: fastValue}},
		ReqBody: outerBody,
	}
	reqBytes, err := asReq.Marshal()
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal FAST AS-REQ: %w", err)
	}
	return c.sendToRealm(c.realm, reqBytes)
}

// processFASTASRep validates an armored AS-REP: it decrypts the armored FAST
// response with the armor key, applies the strengthen-key to derive the reply
// key (RFC 6113 §5.4.3), decrypts the AS-REP enc-part, verifies nonces, and
// stores the TGT.
func (c *KerberosClient) processFASTASRep(resp, armorKey []byte, armorEType, clientEType int, salt string, s2kParams []byte, requestNonce int) error {
	var asRep messages.ASRep
	if _, err := asRep.Unmarshal(resp); err != nil {
		return fmt.Errorf("kerberos: parse FAST AS-REP: %w", err)
	}

	fastResp, err := c.decryptFASTReply(asRep.PAData, armorKey, armorEType)
	if err != nil {
		return fmt.Errorf("kerberos: decrypt FAST reply: %w", err)
	}
	if fastResp.Nonce != requestNonce {
		return fmt.Errorf("kerberos: FAST response nonce mismatch: got %d, want %d", fastResp.Nonce, requestNonce)
	}

	// The AS-REP enc-part is encrypted with the strengthened reply key. MS-KILE
	// requires strengthen-key when the encrypted-challenge factor is used.
	clientKey, err := c.cred.Key(clientEType, salt, s2kParams)
	if err != nil {
		return fmt.Errorf("kerberos: derive client key for reply: %w", err)
	}
	replyKey := clientKey
	replyEType := clientEType
	if fastResp.StrengthenKey != nil {
		sk := fastResp.StrengthenKey
		replyKey, replyEType, err = kerbcrypto.KRBFXCF2(
			sk.KeyValue, sk.KeyType, clientKey, clientEType,
			fastPepperStrengthenKey, fastPepperReplyKey)
		if err != nil {
			return fmt.Errorf("kerberos: strengthen reply key: %w", err)
		}
	}

	encPlain, err := kerbcrypto.Decrypt(replyEType, replyKey, kerbcrypto.KeyUsageASRepEncPart, asRep.EncPart.Cipher)
	if err != nil {
		return fmt.Errorf("kerberos: decrypt FAST AS-REP enc-part: %w", err)
	}
	var encASRep messages.EncASRepPart
	if _, err := encASRep.Unmarshal(encPlain); err != nil {
		return fmt.Errorf("kerberos: parse EncASRepPart: %w", err)
	}
	if encASRep.Nonce != requestNonce {
		return fmt.Errorf("kerberos: FAST AS-REP nonce mismatch: got %d, want %d", encASRep.Nonce, requestNonce)
	}

	c.tgtTicket = asRep.Ticket
	c.tgtTicketRaw = asRep.TicketRaw
	c.sessionKey = encASRep.Key.KeyValue
	c.sessionEType = encASRep.Key.KeyType
	c.tgtEnc = encASRep
	c.hasTGT = true
	return nil
}

// decryptFASTReply locates PA-FX-FAST in a reply's padata, unwraps the armored
// reply, and decrypts the KrbFastResponse with the armor key (key usage 52).
func (c *KerberosClient) decryptFASTReply(paData []messages.PAData, armorKey []byte, armorEType int) (*messages.KrbFastResponse, error) {
	for _, pa := range paData {
		if pa.PADataType != messages.PAFXFast {
			continue
		}
		rep, err := messages.ParsePAFXFastReply(pa.PADataValue)
		if err != nil {
			return nil, err
		}
		plain, err := kerbcrypto.Decrypt(armorEType, armorKey, kerbcrypto.KeyUsageFASTRep, rep.EncFastRep.Cipher)
		if err != nil {
			return nil, fmt.Errorf("decrypt enc-fast-rep: %w", err)
		}
		var fastResp messages.KrbFastResponse
		if _, err := fastResp.Unmarshal(plain); err != nil {
			return nil, fmt.Errorf("parse KrbFastResponse: %w", err)
		}
		return &fastResp, nil
	}
	return nil, fmt.Errorf("no PA-FX-FAST in reply")
}

// parseFASTError inspects an armored KRB-ERROR: it decrypts the FAST reply (if
// present) and extracts the nested PA-FX-ERROR (the real error), a PA-FX-COOKIE
// to echo on retry, and any PA-ETYPE-INFO2 correcting the client key. When the
// error is not FAST-wrapped, all return values are nil.
func (c *KerberosClient) parseFASTError(krbErr messages.KRBError, armorKey []byte, armorEType int) (innerErr *messages.KRBError, cookie *messages.PAData, info messages.ETypeInfo2) {
	if len(krbErr.EData) == 0 {
		return nil, nil, nil
	}
	var methodData []messages.PAData
	if _, err := asn1.Unmarshal(krbErr.EData, &methodData); err != nil {
		return nil, nil, nil
	}
	fastResp, err := c.decryptFASTReply(methodData, armorKey, armorEType)
	if err != nil {
		return nil, nil, nil
	}
	for i := range fastResp.PAData {
		pa := fastResp.PAData[i]
		switch pa.PADataType {
		case messages.PAFXCookie:
			cp := pa
			cookie = &cp
		case messages.PAFXError:
			var e messages.KRBError
			if _, err := e.Unmarshal(pa.PADataValue); err == nil {
				ec := e
				innerErr = &ec
			}
		case messages.PAETypeInfo2:
			var i2 messages.ETypeInfo2
			if _, err := i2.Unmarshal(pa.PADataValue); err == nil && len(i2) > 0 {
				info = i2
			}
		}
	}
	// PA-ETYPE-INFO2 may also ride in the nested error's own EData.
	if info == nil && innerErr != nil && len(innerErr.EData) > 0 {
		var i2 messages.ETypeInfo2
		if _, err := i2.Unmarshal(innerErr.EData); err == nil && len(i2) > 0 {
			info = i2
		}
	}
	return innerErr, cookie, info
}
