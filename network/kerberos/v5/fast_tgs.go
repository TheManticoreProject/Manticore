package kerberos

import (
	"crypto/rand"
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// This file extends RFC 6113 FAST (Kerberos armoring) to the TGS exchange, as a
// companion to the AS-exchange path in fast.go.
//
// The TGS armor is not a separate FX_FAST_ARMOR_AP_REQUEST. RFC 6113 §5.4.2
// specifies that when the ticket presented in the PA-TGS-REQ authenticator is a
// TGT, the client omits the KrbFastArmoredReq armor field entirely and instead
// places a subkey in that PA-TGS-REQ authenticator. The armor key is then
// computed exactly as for FX_FAST_ARMOR_AP_REQUEST (RFC 6113 §5.4.1.1) —
// KRB-FX-CF2(subkey, ticket-session-key, "subkeyarmor", "ticketarmor") — but the
// "ticket" is the TGT being presented in the TGS-REQ and the subkey is the one
// in its AP-REQ authenticator.
//
// Two further TGS-specific rules distinguish this from the AS path:
//   - the req-checksum (key usage 50) is computed over the AP-REQ in the
//     PA-TGS-REQ padata, not over the outer KDC-REQ-BODY (RFC 6113 §5.4.2); and
//   - the KDC MUST return a strengthen-key in the TGS reply, so the reply key is
//     always KRB-FX-CF2(strengthen-key, subkey, "strengthenkey", "replykey")
//     and the reply enc-part is keyed by the sub-session key (usage 9)
//     (RFC 6113 §5.4.3).

// fastTGSContext carries the per-request FAST state a TGS reply needs to be
// unwrapped: the armor key (to decrypt the KrbFastResponse), the authenticator
// subkey (the base reply key the strengthen-key is combined with), and the inner
// request nonce the reply must echo.
type fastTGSContext struct {
	armorKey    []byte
	armorEType  int
	subkey      []byte
	subkeyEType int
	nonce       int
}

// buildTGSAPReqWithSubkey builds the PA-TGS-REQ AP-REQ that presents tgt, with
// the given subkey placed in the authenticator (RFC 6113 §5.4.2 requires the
// subkey for TGS armoring). The authenticator is encrypted with the TGT session
// key under key usage 7 (the PA-TGS-REQ authenticator usage).
func (c *KerberosClient) buildTGSAPReqWithSubkey(tgt messages.Ticket, tgtRaw, sessionKey []byte, sessionEType int, subkey []byte) ([]byte, error) {
	now := c.now()

	var seqBuf [4]byte
	if _, err := rand.Read(seqBuf[:]); err != nil {
		return nil, err
	}
	seqNum := int(binary.BigEndian.Uint32(seqBuf[:]) & 0x7fffffff)

	auth := &messages.Authenticator{
		AVno:      messages.KerberosV5,
		CRealm:    c.realm,
		CName:     messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{c.username}},
		CUSec:     now.Nanosecond() / 1000,
		CTime:     now,
		SubKey:    &messages.EncryptionKey{KeyType: sessionEType, KeyValue: subkey},
		SeqNumber: seqNum,
	}
	authBytes, err := auth.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal TGS Authenticator: %w", err)
	}
	encAuth, err := kerbcrypto.Encrypt(sessionEType, sessionKey, kerbcrypto.KeyUsageTGSReqPAAPReqAuthen, authBytes)
	if err != nil {
		return nil, fmt.Errorf("encrypt TGS Authenticator: %w", err)
	}

	apReq := &messages.APReq{
		PVNO:      messages.KerberosV5,
		MsgType:   messages.MsgTypeAPReq,
		APOptions: asn1.BitString{Bytes: []byte{0x00, 0x00, 0x00, 0x00}, BitLength: 32},
		Ticket:    tgt,
		TicketRaw: tgtRaw,
		Authenticator: messages.EncryptedData{
			EType:  sessionEType,
			Cipher: encAuth,
		},
	}
	return apReq.Marshal()
}

// tgsReqBody builds the KDC-REQ-BODY for a service-ticket request toward sname in
// bodyRealm. It mirrors the body assembled by the non-FAST tgsExchange so the
// armored (inner) request the KDC honors is identical in shape.
func (c *KerberosClient) tgsReqBody(bodyRealm string, sname messages.PrincipalName, nonce int) messages.KDCReqBody {
	return messages.KDCReqBody{
		KDCOptions: kdcOptionsForTGSReq(),
		Realm:      bodyRealm,
		SName:      sname,
		Till:       c.now().Add(24 * time.Hour),
		Nonce:      nonce,
		EType:      c.serviceTicketETypes(),
	}
}

// buildFASTTGSReq assembles a FAST-armored TGS-REQ for sname at the current hop,
// presenting tgt as both the ticket-granting ticket and (per RFC 6113 §5.4.2) the
// FAST armor. It returns the marshaled request and the fastTGSContext the reply
// processor needs. cookie, when non-nil, is echoed as PA-FX-COOKIE in the armored
// (inner) padata.
func (c *KerberosClient) buildFASTTGSReq(
	bodyRealm string,
	sname messages.PrincipalName,
	includePAC bool,
	tgt messages.Ticket, tgtRaw, sessionKey []byte, sessionEType int,
	cookie *messages.PAData,
) ([]byte, *fastTGSContext, error) {

	// A fresh subkey carried in the PA-TGS-REQ authenticator; combined with the
	// TGT session key it yields the armor key (RFC 6113 §5.4.1.1 / §5.4.2). Both
	// inputs share the TGT session etype, so the armor key adopts that etype.
	subkey := make([]byte, kerbcrypto.KeyLen(sessionEType))
	if _, err := rand.Read(subkey); err != nil {
		return nil, nil, fmt.Errorf("kerberos: FAST TGS subkey: %w", err)
	}
	armorKey, armorEType, err := kerbcrypto.KRBFXCF2(
		subkey, sessionEType,
		sessionKey, sessionEType,
		fastPepperSubkeyArmor, fastPepperTicketArmor)
	if err != nil {
		return nil, nil, fmt.Errorf("kerberos: derive FAST TGS armor key: %w", err)
	}

	// The AP-REQ carrying the subkey. It goes verbatim into the outer PA-TGS-REQ
	// and is also the object the req-checksum covers (RFC 6113 §5.4.2).
	apReqBytes, err := c.buildTGSAPReqWithSubkey(tgt, tgtRaw, sessionKey, sessionEType, subkey)
	if err != nil {
		return nil, nil, fmt.Errorf("kerberos: build FAST TGS AP-REQ: %w", err)
	}

	// Inner (armor-protected) padata + body. PA-PAC-REQUEST rides inside the armor,
	// as does the FX-COOKIE to echo when the KDC handed one back. The KDC uses this
	// inner req-body in preference to the outer, unprotected one.
	pacBool := byte(0xff)
	if !includePAC {
		pacBool = 0x00
	}
	innerPAData := []messages.PAData{
		{PADataType: messages.PAPACRequest, PADataValue: []byte{0x30, 0x05, 0xa0, 0x03, 0x01, 0x01, pacBool}},
	}
	if cookie != nil {
		innerPAData = append(innerPAData, *cookie)
	}
	nonce := randomNonce()
	innerBody := c.tgsReqBody(bodyRealm, sname, nonce)

	fastReq := &messages.KrbFastReq{
		FastOptions: messages.NewKerberosFlags(),
		PAData:      innerPAData,
		ReqBody:     innerBody,
	}
	fastReqBytes, err := fastReq.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("kerberos: marshal KrbFastReq (TGS): %w", err)
	}
	encFastReq, err := kerbcrypto.Encrypt(armorEType, armorKey, kerbcrypto.KeyUsageFASTEnc, fastReqBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("kerberos: encrypt KrbFastReq (TGS): %w", err)
	}

	// req-checksum: a keyed checksum over the AP-REQ, computed with the armor key
	// (RFC 6113 §5.4.2, key usage 50). The TGS case checksums the AP-REQ, not the
	// KDC-REQ-BODY as the AS case does.
	cksumType, ok := kerbcrypto.ChecksumTypeForEType(armorEType)
	if !ok {
		return nil, nil, fmt.Errorf("kerberos: no checksum type for armor etype %d", armorEType)
	}
	cksum, err := kerbcrypto.GetChecksum(cksumType, armorKey, kerbcrypto.KeyUsageFASTReqChksum, apReqBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("kerberos: FAST TGS req-checksum: %w", err)
	}

	// No armor field for the TGS case (RFC 6113 §5.4.2): the armor is the TGT
	// itself, presented in the PA-TGS-REQ AP-REQ.
	armoredReq := &messages.KrbFastArmoredReq{
		ReqChecksum: messages.Checksum{CKSumType: cksumType, Checksum: cksum},
		EncFastReq:  messages.EncryptedData{EType: armorEType, Cipher: encFastReq},
	}
	fastValue, err := messages.MarshalPAFXFastRequest(armoredReq)
	if err != nil {
		return nil, nil, fmt.Errorf("kerberos: marshal PA-FX-FAST (TGS): %w", err)
	}

	// The outer TGS-REQ carries the PA-TGS-REQ (so the KDC can locate the armor
	// ticket) and PA-FX-FAST. Its body is present but ignored in favor of the
	// armored inner body, so it uses its own throwaway nonce.
	tgsReq := &messages.TGSReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeTGSReq,
		PAData: []messages.PAData{
			{PADataType: messages.PATGSReq, PADataValue: apReqBytes},
			{PADataType: messages.PAFXFast, PADataValue: fastValue},
		},
		ReqBody: c.tgsReqBody(bodyRealm, sname, randomNonce()),
	}
	reqBytes, err := tgsReq.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("kerberos: marshal FAST TGS-REQ: %w", err)
	}

	ctx := &fastTGSContext{
		armorKey:    armorKey,
		armorEType:  armorEType,
		subkey:      subkey,
		subkeyEType: sessionEType,
		nonce:       nonce,
	}
	return reqBytes, ctx, nil
}

// processFASTTGSRep parses a FAST-armored TGS reply. On a TGS-REP it decrypts the
// KrbFastResponse with the armor key (usage 52), derives the strengthened reply
// key (RFC 6113 §5.4.3 mandates strengthen-key for TGS) from the authenticator
// subkey, decrypts the TGS-REP enc-part with it (usage 9, the sub-session key),
// and returns the parsed ticket and enc-part. A KRB-ERROR is returned as
// *messages.KRBError (with a nil error) so the caller can react to a recoverable
// code; any FAST-armored inner error is preferred and an FX-COOKIE to echo on a
// retry is surfaced.
func (c *KerberosClient) processFASTTGSRep(resp []byte, ctx *fastTGSContext) (*messages.TGSRep, *messages.EncTGSRepPart, *messages.KRBError, *messages.PAData, error) {
	var krbErr messages.KRBError
	if _, parseErr := krbErr.Unmarshal(resp); parseErr == nil {
		innerErr, cookie, _ := c.parseFASTError(krbErr, ctx.armorKey, ctx.armorEType)
		effective := krbErr
		if innerErr != nil {
			effective = *innerErr
		}
		return nil, nil, &effective, cookie, nil
	}

	var tgsRep messages.TGSRep
	if _, err := tgsRep.Unmarshal(resp); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("kerberos: parse FAST TGS-REP: %w", err)
	}

	fastResp, err := c.decryptFASTReply(tgsRep.PAData, ctx.armorKey, ctx.armorEType)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("kerberos: decrypt FAST TGS reply: %w", err)
	}
	if fastResp.Nonce != ctx.nonce {
		return nil, nil, nil, nil, fmt.Errorf("kerberos: FAST TGS response nonce mismatch: got %d, want %d", fastResp.Nonce, ctx.nonce)
	}

	// The base reply key is the authenticator sub-session key; the KDC MUST
	// strengthen it for TGS (RFC 6113 §5.4.3). The strengthened key replaces the
	// key value but not the usage: a subkey was sent, so the enc-part is keyed by
	// the sub-session key (usage 9).
	replyKey := ctx.subkey
	replyEType := ctx.subkeyEType
	if fastResp.StrengthenKey != nil {
		sk := fastResp.StrengthenKey
		replyKey, replyEType, err = kerbcrypto.KRBFXCF2(
			sk.KeyValue, sk.KeyType, ctx.subkey, ctx.subkeyEType,
			fastPepperStrengthenKey, fastPepperReplyKey)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("kerberos: strengthen TGS reply key: %w", err)
		}
	}

	encPlain, err := kerbcrypto.Decrypt(replyEType, replyKey, kerbcrypto.KeyUsageTGSRepEncSubSessionKey, tgsRep.EncPart.Cipher)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("kerberos: decrypt FAST TGS-REP enc-part: %w", err)
	}
	var encRep messages.EncTGSRepPart
	if _, err := encRep.Unmarshal(encPlain); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("kerberos: parse EncTGSRepPart: %w", err)
	}
	if encRep.Nonce != ctx.nonce {
		return nil, nil, nil, nil, fmt.Errorf("kerberos: FAST TGS-REP nonce mismatch: got %d, want %d", encRep.Nonce, ctx.nonce)
	}
	return &tgsRep, &encRep, nil, nil, nil
}

// tgsExchangeFAST performs a single FAST-armored TGS-REQ/REP round-trip against
// the given KDC endpoints, presenting tgt as both the ticket-granting ticket and
// the FAST armor. It mirrors the return contract of the non-FAST tgsExchange: a
// KRB-ERROR is returned as *messages.KRBError (nil error) so the chase loop can
// react to WRONG_REALM / KRB_AP_ERR_SKEW. A single FX-COOKIE handed back inside a
// FAST-armored error is echoed on one retry (RFC 6113 §5.4.4.3).
func (c *KerberosClient) tgsExchangeFAST(
	bodyRealm string,
	endpoints []kdcEndpoint,
	sname messages.PrincipalName,
	includePAC bool,
	tgt messages.Ticket, tgtRaw, sessionKey []byte, sessionEType int,
) (*messages.TGSRep, *messages.EncTGSRepPart, *messages.KRBError, error) {

	var cookie *messages.PAData
	for attempt := 0; attempt < 2; attempt++ {
		reqBytes, ctx, err := c.buildFASTTGSReq(bodyRealm, sname, includePAC, tgt, tgtRaw, sessionKey, sessionEType, cookie)
		if err != nil {
			return nil, nil, nil, err
		}
		resp, err := kdcSendEndpoints(c.resolver, endpoints, reqBytes)
		if err != nil {
			return nil, nil, nil, err
		}

		rep, encRep, krbErr, newCookie, err := c.processFASTTGSRep(resp, ctx)
		if err != nil {
			return nil, nil, nil, err
		}
		if krbErr != nil {
			// A cookie handed back on the first attempt is echoed once; any other
			// error (WRONG_REALM, SKEW, hard failure) is handed to the chase loop.
			if newCookie != nil && cookie == nil && attempt == 0 {
				cookie = newCookie
				continue
			}
			return nil, nil, krbErr, nil
		}
		return rep, encRep, nil, nil
	}
	return nil, nil, nil, fmt.Errorf("kerberos: FAST TGS: retry budget exhausted")
}
