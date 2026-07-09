package kerberos

import (
	"fmt"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// buildRenewalTGSReq constructs the TGS-REQ for a ticket RENEW or VALIDATE
// exchange (RFC 4120 §3.3.3). Both are ordinary ticket-granting-service requests
// that present the client's existing TGT via a PA-TGS-REQ AP-REQ, target the
// ticket-granting service (krbtgt/REALM), and set exactly one distinguishing KDC
// option — renew (bit 30) or validate (bit 31) — alongside the renewable flag so
// the reissued ticket keeps its renewable window. The ticket to be renewed or
// validated is the one carried in the AP-REQ (padata), not additional-tickets.
//
// It is factored out of Renew/Validate so the request shape can be verified
// without a live KDC (see renew_test.go).
func (c *KerberosClient) buildRenewalTGSReq(option, nonce int) (*messages.TGSReq, error) {
	apReqBytes, err := c.buildAPReq()
	if err != nil {
		return nil, fmt.Errorf("kerberos: build AP-REQ: %w", err)
	}

	// The renewed/validated ticket is a TGT, so the requested server is the
	// ticket-granting service of the client's own realm.
	sname := messages.PrincipalName{
		NameType:   messages.NameTypeSRVInst,
		NameString: []string{"krbtgt", c.realm},
	}

	// Request an endtime as far out as the renewable window allows. RFC 4120
	// §3.3.3: the KDC caps the new endtime at the ticket's renew-till (and its
	// original lifetime), so asking for renew-till yields the maximal fresh
	// endtime without over-reaching. Fall back to a 24h window when the stored
	// TGT carries no renew-till (e.g. a validation of a non-renewable ticket).
	till := c.now().Add(24 * time.Hour)
	if option == kdcOptionRenew && !c.tgtEnc.RenewTill.IsZero() {
		till = c.tgtEnc.RenewTill
	}

	return &messages.TGSReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeTGSReq,
		PAData: []messages.PAData{
			{PADataType: messages.PATGSReq, PADataValue: apReqBytes},
		},
		ReqBody: messages.KDCReqBody{
			KDCOptions: encodeKDCOptions(kdcOptionRenewable, option),
			Realm:      c.realm,
			SName:      sname,
			Till:       till,
			Nonce:      nonce,
			EType:      c.serviceTicketETypes(),
		},
	}, nil
}

// Renew refreshes the client's renewable Ticket Granting Ticket without
// re-authenticating. It sends a TGS-REQ with the renew KDC option, presenting the
// current TGT in a PA-TGS-REQ AP-REQ (RFC 4120 §3.3.3). The KDC returns a fresh
// TGT with a new session key and advanced start/end times (capped at the
// ticket's renew-till), which replaces the client's stored TGT and session key on
// success. GetTGT must have succeeded first, and the current TGT must be
// renewable and not past its renew-till, or the KDC rejects the request.
func (c *KerberosClient) Renew() error {
	return c.renewOrValidate(kdcOptionRenew, "renew")
}

// Validate turns a postdated (INVALID-flagged) TGT into a usable one once its
// start time has passed. It sends a TGS-REQ with the validate KDC option,
// presenting the ticket in a PA-TGS-REQ AP-REQ (RFC 4120 §3.3.3); the KDC clears
// the INVALID flag and returns a fresh ticket, which replaces the client's stored
// TGT and session key on success. GetTGT must have succeeded first. Many KDCs
// (including default Active Directory policy) disable postdating, in which case
// the KDC rejects the request.
func (c *KerberosClient) Validate() error {
	return c.renewOrValidate(kdcOptionValidate, "validate")
}

// renewOrValidate performs the RENEW/VALIDATE TGS exchange and, on success,
// installs the reissued TGT (ticket, raw bytes, session key and enc-part) on the
// client. A single KRB_AP_ERR_SKEW reply triggers a clock realignment and one
// retry (RFC 4120 §5.9.1), mirroring the other exchanges.
func (c *KerberosClient) renewOrValidate(option int, label string) error {
	if !c.hasTGT {
		return fmt.Errorf("kerberos: no TGT: call GetTGT first")
	}

	for attempt := 0; attempt < 2; attempt++ {
		nonce := randomNonce()
		req, err := c.buildRenewalTGSReq(option, nonce)
		if err != nil {
			return err
		}
		reqBytes, err := req.Marshal()
		if err != nil {
			return fmt.Errorf("kerberos: marshal %s TGS-REQ: %w", label, err)
		}
		resp, err := c.sendToRealm(c.realm, reqBytes)
		if err != nil {
			return err
		}

		var krbErr messages.KRBError
		if _, parseErr := krbErr.Unmarshal(resp); parseErr == nil {
			if krbErr.ErrorCode == messages.ErrSkew && attempt == 0 && c.applyClockSkew(krbErr) {
				continue // realign to the KDC clock and retry once
			}
			return fmt.Errorf("kerberos: %s error %d: %s", label, krbErr.ErrorCode, krbErr.EText)
		}

		var tgsRep messages.TGSRep
		if _, err := tgsRep.Unmarshal(resp); err != nil {
			return fmt.Errorf("kerberos: parse %s TGS-REP: %w", label, err)
		}

		// The reply enc-part is sealed under the presented TGT's session key.
		encPlain, err := kerbcrypto.Decrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageTGSRepEncSessionKey, tgsRep.EncPart.Cipher)
		if err != nil {
			return fmt.Errorf("kerberos: decrypt %s TGS-REP enc-part: %w", label, err)
		}
		var encRep messages.EncTGSRepPart
		if _, err := encRep.Unmarshal(encPlain); err != nil {
			return fmt.Errorf("kerberos: parse %s EncTGSRepPart: %w", label, err)
		}
		if encRep.Nonce != nonce {
			return fmt.Errorf("kerberos: %s nonce mismatch: got %d, want %d", label, encRep.Nonce, nonce)
		}

		c.storeReissuedTGT(&tgsRep, &encRep)
		return nil
	}
	return fmt.Errorf("kerberos: %s: clock-skew retry exhausted", label)
}

// storeReissuedTGT replaces the client's cached TGT with the ticket returned by a
// RENEW/VALIDATE exchange: the new ticket and its raw bytes, the new session key
// and etype, and the refreshed ticket times/flags. Subsequent GetTGS calls then
// use the reissued TGT.
func (c *KerberosClient) storeReissuedTGT(rep *messages.TGSRep, enc *messages.EncTGSRepPart) {
	c.tgtTicket = rep.Ticket
	c.tgtTicketRaw = rep.TicketRaw
	c.sessionKey = enc.Key.KeyValue
	c.sessionEType = enc.Key.KeyType
	c.tgtEnc = messages.EncASRepPart{
		Key:       enc.Key,
		Nonce:     enc.Nonce,
		Flags:     enc.Flags,
		AuthTime:  enc.AuthTime,
		StartTime: enc.StartTime,
		EndTime:   enc.EndTime,
		RenewTill: enc.RenewTill,
		SRealm:    enc.SRealm,
		SName:     enc.SName,
	}
}
