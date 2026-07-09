package kerberos

import (
	"encoding/asn1"
	"fmt"
	"strings"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// maxReferralHops bounds how many cross-realm referral TGTs the client will
// chase before giving up. A transitive trust path is at most a few realms deep;
// this ceiling protects against a misbehaving KDC that keeps issuing referrals.
const maxReferralHops = 10

// WithRealmKDC registers the KDC host to contact when the cross-realm referral
// chase reaches the given realm. The realm is uppercased automatically. This is
// the minimal, no-dependency way to resolve a target-realm KDC; automatic
// discovery (DNS SRV) is used for any realm not registered here. Returns the
// client to allow fluent chaining.
func (c *KerberosClient) WithRealmKDC(realm, kdcHost string) *KerberosClient {
	if c.realmKDCs == nil {
		c.realmKDCs = make(map[string]string)
	}
	c.realmKDCs[strings.ToUpper(realm)] = kdcHost
	return c
}

// WithKDCResolver installs a custom function that resolves a realm to a KDC host
// for the cross-realm referral chase. It takes precedence over DNS-SRV discovery
// but not over explicit WithRealmKDC entries or the client's own home realm.
// Returns the client to allow fluent chaining.
func (c *KerberosClient) WithKDCResolver(fn func(realm string) (string, error)) *KerberosClient {
	c.kdcResolver = fn
	return c
}

// resolveKDCForRealm returns the primary KDC host to contact for the given realm
// during referral chasing. Resolution order: explicit WithRealmKDC overrides,
// then the client's own configured KDC for its home realm, then a custom
// resolver, then DNS-SRV discovery (see endpointsForRealm). The full failover
// list is available via endpointsForRealm; this helper exposes just the first
// (highest-priority) host for callers that want a single name.
func (c *KerberosClient) resolveKDCForRealm(realm string) (string, error) {
	endpoints, err := c.endpointsForRealm(realm)
	if err != nil {
		return "", err
	}
	return endpoints[0].host, nil
}

// chaseServiceTicket obtains a service ticket for sname, following any
// cross-realm referral TGTs the KDC returns (RFC 6806, MS-KILE Section 3.3.5.1).
//
// Starting from the client's home TGT/KDC, it issues a TGS-REQ toward the
// current realm. If the reply is the requested service ticket it is returned. If
// the reply is a referral TGT (server name krbtgt/NEXT-REALM), the chase switches
// to NEXT-REALM: it resolves that realm's KDC, presents the referral TGT, and
// re-issues the TGS-REQ — repeating until the service ticket is obtained. A
// KDC_ERR_WRONG_REALM error is handled by retrying with the corrected target
// realm the KDC named. Cycles and over-long chains are rejected.
func (c *KerberosClient) chaseServiceTicket(sname messages.PrincipalName, includePAC bool) (messages.Ticket, []byte, []byte, int, error) {
	// The credential presented at the current hop. It starts as the home TGT and
	// becomes each successive referral TGT.
	tgt := c.tgtTicket
	tgtRaw := c.tgtTicketRaw
	sessionKey := c.sessionKey
	sessionEType := c.sessionEType

	// bodyRealm is the KDC-REQ-BODY realm — the realm whose KDC we are asking.
	bodyRealm := c.realm
	// endpoints are the failover-ordered KDCs serving bodyRealm.
	endpoints, err := c.endpointsForRealm(bodyRealm)
	if err != nil {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: resolve KDC for realm %q: %w", bodyRealm, err)
	}

	// Realms whose KDC we have already presented a TGT to, to detect referral
	// cycles (an authority realm must not be revisited).
	visited := map[string]bool{strings.ToUpper(bodyRealm): true}
	// Corrected realms we have already retried for WRONG_REALM, to avoid a
	// ping-pong between two KDCs that keep correcting each other. Kept separate
	// from visited so the follow-up referral toward the corrected realm is not
	// itself mistaken for a cycle.
	wrongRealmTried := map[string]bool{}
	// skewRetried guards the single KRB_AP_ERR_SKEW retry.
	skewRetried := false

	for hop := 0; hop < maxReferralHops; hop++ {
		rep, encRep, krbErr, err := c.tgsExchange(bodyRealm, endpoints, sname, includePAC, tgt, tgtRaw, sessionKey, sessionEType)
		if err != nil {
			return messages.Ticket{}, nil, nil, 0, err
		}
		if krbErr != nil {
			// KDC_ERR_WRONG_REALM: the KDC tells the client the realm the
			// principal actually lives in (RFC 4120 Section 3.3.3.1). Retry once
			// per corrected realm against the same (home) KDC so it can issue a
			// referral toward that realm.
			if krbErr.ErrorCode == messages.ErrWrongRealm && krbErr.CRealm != "" {
				corrected := strings.ToUpper(krbErr.CRealm)
				if !wrongRealmTried[corrected] {
					wrongRealmTried[corrected] = true
					bodyRealm = corrected
					continue
				}
			}
			// KRB_AP_ERR_SKEW: align our clock to the KDC's and retry the same hop
			// once (RFC 4120 Section 5.9.1).
			if krbErr.ErrorCode == messages.ErrSkew && !skewRetried && c.applyClockSkew(*krbErr) {
				skewRetried = true
				continue
			}
			return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: TGS error %d: %s", krbErr.ErrorCode, krbErr.EText)
		}

		nextRealm, isReferral := referralTargetRealm(rep, sname)
		if !isReferral {
			// The requested service ticket was issued — done.
			return rep.Ticket, rep.TicketRaw, encRep.Key.KeyValue, encRep.Key.KeyType, nil
		}

		nextRealm = strings.ToUpper(nextRealm)
		if visited[nextRealm] {
			return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: cross-realm referral cycle detected at realm %q", nextRealm)
		}
		nextEndpoints, err := c.endpointsForRealm(nextRealm)
		if err != nil {
			return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: resolve KDC for referral realm %q: %w", nextRealm, err)
		}

		// Advance to the next realm, presenting the referral TGT there.
		visited[nextRealm] = true
		bodyRealm = nextRealm
		endpoints = nextEndpoints
		tgt = rep.Ticket
		tgtRaw = rep.TicketRaw
		sessionKey = encRep.Key.KeyValue
		sessionEType = encRep.Key.KeyType
	}

	return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: too many cross-realm referrals (>%d) chasing %s", maxReferralHops, strings.Join(sname.NameString, "/"))
}

// tgsExchange performs a single TGS-REQ/REP round-trip against the given KDC
// endpoints (failing over across them), using bodyRealm as the KDC-REQ-BODY
// realm and presenting the supplied ticket-granting ticket. It returns the
// parsed TGS-REP with its decrypted enc-part on success. If the KDC answers with
// a KRB-ERROR it is returned as *messages.KRBError (with a nil error) so the
// caller can react to WRONG_REALM or KRB_AP_ERR_SKEW.
func (c *KerberosClient) tgsExchange(
	bodyRealm string,
	endpoints []kdcEndpoint,
	sname messages.PrincipalName,
	includePAC bool,
	tgt messages.Ticket, tgtRaw, sessionKey []byte, sessionEType int,
) (*messages.TGSRep, *messages.EncTGSRepPart, *messages.KRBError, error) {

	apReqBytes, err := c.buildAPReqWith(tgt, tgtRaw, sessionKey, sessionEType)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("kerberos: build AP-REQ: %w", err)
	}

	nonce := randomNonce()
	// PA-PAC-REQUEST: SEQUENCE { [0] BOOLEAN } — TRUE=0xff, FALSE=0x00.
	pacBool := byte(0xff)
	if !includePAC {
		pacBool = 0x00
	}
	tgsReq := &messages.TGSReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeTGSReq,
		PAData: []messages.PAData{
			{PADataType: messages.PATGSReq, PADataValue: apReqBytes},
			{PADataType: messages.PAPACRequest, PADataValue: []byte{0x30, 0x05, 0xa0, 0x03, 0x01, 0x01, pacBool}},
		},
		ReqBody: messages.KDCReqBody{
			KDCOptions: kdcOptionsForTGSReq(),
			Realm:      bodyRealm,
			SName:      sname,
			Till:       c.now().Add(24 * time.Hour),
			Nonce:      nonce,
			EType:      c.serviceTicketETypes(),
		},
	}

	tgsReqBytes, err := tgsReq.Marshal()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("kerberos: marshal TGS-REQ: %w", err)
	}
	resp, err := kdcSendEndpoints(c.resolver, endpoints, tgsReqBytes)
	if err != nil {
		return nil, nil, nil, err
	}

	// A KRB-ERROR is handed back to the caller (it may be a recoverable
	// WRONG_REALM) rather than turned into an error here.
	var krbErr messages.KRBError
	if _, parseErr := krbErr.Unmarshal(resp); parseErr == nil {
		return nil, nil, &krbErr, nil
	}

	var tgsRep messages.TGSRep
	if _, err := tgsRep.Unmarshal(resp); err != nil {
		return nil, nil, nil, fmt.Errorf("kerberos: parse TGS-REP: %w", err)
	}

	// The reply enc-part is encrypted under the TGT session key of the ticket we
	// presented at this hop (the referral TGT's key on cross-realm hops).
	encPlain, err := kerbcrypto.Decrypt(sessionEType, sessionKey, kerbcrypto.KeyUsageTGSRepEncSessionKey, tgsRep.EncPart.Cipher)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("kerberos: decrypt TGS-REP enc-part: %w", err)
	}

	var encRep messages.EncTGSRepPart
	if _, err := encRep.Unmarshal(encPlain); err != nil {
		return nil, nil, nil, fmt.Errorf("kerberos: parse EncTGSRepPart: %w", err)
	}

	// RFC 4120 Section 3.3.3: the reply nonce must match the request nonce.
	// Rejecting a mismatch defends against replays of captured TGS-REPs that
	// happen to decrypt under the presented session key.
	if encRep.Nonce != nonce {
		return nil, nil, nil, fmt.Errorf("kerberos: TGS-REP nonce mismatch: got %d, want %d", encRep.Nonce, nonce)
	}

	return &tgsRep, &encRep, nil, nil
}

// referralTargetRealm decides whether a TGS-REP is a cross-realm referral rather
// than the requested service ticket, and if so returns the realm to chase next.
//
// A referral is signalled by a returned ticket whose server name is the two-part
// TGS principal "krbtgt/NEXT-REALM" while the client asked for a different
// (non-krbtgt) service (RFC 6806 Section 7, MS-KILE Section 3.3.5.1). When the
// KDC additionally supplies a PA-SVR-REFERRAL-INFO element naming the true realm
// (RFC 6806 Section 3.3.3.1) it is preferred, as it is the authoritative hint.
func referralTargetRealm(rep *messages.TGSRep, requested messages.PrincipalName) (string, bool) {
	realm, isReferral := referralRealmFromSName(rep.Ticket.SName, requested)
	if !isReferral {
		return "", false
	}
	if trueRealm, ok := svrReferralRealm(rep.PAData); ok {
		return trueRealm, true
	}
	return realm, true
}

// referralRealmFromSName returns the referral target realm encoded in a returned
// ticket's server name (krbtgt/REALM), or ok=false when the ticket is not a
// referral TGT. If the caller explicitly requested that krbtgt principal, it is
// treated as the requested service — not a referral.
func referralRealmFromSName(ticketSName, requested messages.PrincipalName) (string, bool) {
	if len(ticketSName.NameString) != 2 {
		return "", false
	}
	if !strings.EqualFold(ticketSName.NameString[0], "krbtgt") {
		return "", false
	}
	if principalNameEqualFold(ticketSName, requested) {
		return "", false
	}
	return ticketSName.NameString[1], true
}

// svrReferralData is the PA-SVR-REFERRAL-INFO payload (PA-SVR-REFERRAL-DATA) of
// RFC 6806 Appendix A / MS-KILE Section 2.2.1: the true realm a canonicalized
// server principal belongs to, optionally with the referred name. Field order
// mirrors the ASN.1 module (referred-name [1] precedes referred-realm [0]); in
// practice only referred-realm is sent in TGS replies.
type svrReferralData struct {
	ReferredName  messages.PrincipalName `asn1:"explicit,tag:1,optional"`
	ReferredRealm string                 `asn1:"explicit,tag:0,generalstring"`
}

// svrReferralRealm extracts the true realm from a PA-SVR-REFERRAL-INFO element in
// the given PA-DATA list, if present.
func svrReferralRealm(paList []messages.PAData) (string, bool) {
	for _, pa := range paList {
		if pa.PADataType != messages.PASvrReferralInfo {
			continue
		}
		var d svrReferralData
		if _, err := asn1.Unmarshal(pa.PADataValue, &d); err == nil && d.ReferredRealm != "" {
			return strings.ToUpper(d.ReferredRealm), true
		}
	}
	return "", false
}

// principalNameEqualFold reports whether two principal names have the same
// components, compared case-insensitively (realm and service names are
// case-insensitive in AD practice).
func principalNameEqualFold(a, b messages.PrincipalName) bool {
	if len(a.NameString) != len(b.NameString) {
		return false
	}
	for i := range a.NameString {
		if !strings.EqualFold(a.NameString[i], b.NameString[i]) {
			return false
		}
	}
	return true
}
