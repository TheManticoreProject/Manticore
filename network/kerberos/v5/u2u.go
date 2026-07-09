package kerberos

import (
	"fmt"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// buildU2UTGSReq constructs the TGS-REQ for a user-to-user exchange: a normal
// PA-TGS-REQ (AP-REQ over the client's TGT), the ENC-TKT-IN-SKEY KDC option, the
// target user's TGT in additional-tickets, and the target user as sname. It is
// separated from GetTGSU2U so the request shape can be tested without a KDC.
func (c *KerberosClient) buildU2UTGSReq(targetUser, targetRealm string, targetTGTRaw []byte, nonce int) (*messages.TGSReq, error) {
	apReqBytes, err := c.buildAPReq()
	if err != nil {
		return nil, fmt.Errorf("kerberos: build AP-REQ: %w", err)
	}
	sname := messages.PrincipalName{
		NameType:   messages.NameTypePrincipal,
		NameString: []string{targetUser},
	}
	realm := c.realm
	if targetRealm != "" {
		realm = targetRealm
	}
	return &messages.TGSReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeTGSReq,
		PAData: []messages.PAData{
			{PADataType: messages.PATGSReq, PADataValue: apReqBytes},
		},
		ReqBody: messages.KDCReqBody{
			KDCOptions: encodeKDCOptions(
				kdcOptionForwardable,
				kdcOptionRenewable,
				kdcOptionCanonicalize,
				kdcOptionEncTktInSKey,
			),
			Realm: realm,
			SName: sname,
			Till:  time.Now().UTC().Add(24 * time.Hour),
			Nonce: nonce,
			EType: []int{
				messages.ETypeAES256CTSHMACSHA196,
				messages.ETypeAES128CTSHMACSHA196,
				messages.ETypeRC4HMAC,
			},
			AdditTicketsRaw: [][]byte{targetTGTRaw},
		},
	}, nil
}

// GetTGSU2U performs a user-to-user (U2U) TGS exchange: it requests a service
// ticket to the target user (targetUser in targetRealm, or the client's realm
// if empty), presenting the target user's TGT (targetTGTRaw, raw APPLICATION[1]
// bytes) in additional-tickets with the ENC-TKT-IN-SKEY option. The KDC issues a
// ticket whose enc-part is encrypted with the session key of the target's TGT
// (not a long-term key), which the target verifies via an AP-REQ carrying the
// USE-SESSION-KEY option.
//
// GetTGT must have succeeded first (for the client's own TGT). Returns the
// service ticket, its raw bytes, and the ticket session key.
func (c *KerberosClient) GetTGSU2U(targetUser, targetRealm string, targetTGTRaw []byte) (messages.Ticket, []byte, []byte, error) {
	if !c.hasTGT {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: no TGT: call GetTGT first")
	}
	if targetUser == "" {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: U2U requires a target user")
	}
	if len(targetTGTRaw) == 0 {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: U2U requires the target user's TGT")
	}

	nonce := randomNonce()
	tgsReq, err := c.buildU2UTGSReq(targetUser, targetRealm, targetTGTRaw, nonce)
	if err != nil {
		return messages.Ticket{}, nil, nil, err
	}

	tgsReqBytes, err := tgsReq.Marshal()
	if err != nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: marshal U2U TGS-REQ: %w", err)
	}
	resp, err := kdcSend(c.kdcHost, defaultKDCPort, tgsReqBytes)
	if err != nil {
		return messages.Ticket{}, nil, nil, err
	}

	var krbErr messages.KRBError
	if _, parseErr := krbErr.Unmarshal(resp); parseErr == nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: U2U error %d: %s", krbErr.ErrorCode, krbErr.EText)
	}

	var tgsRep messages.TGSRep
	if _, err := tgsRep.Unmarshal(resp); err != nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: parse U2U TGS-REP: %w", err)
	}

	encPlain, err := kerbcrypto.Decrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageTGSRepEncSessionKey, tgsRep.EncPart.Cipher)
	if err != nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: decrypt U2U TGS-REP enc-part: %w", err)
	}
	var encTGSRep messages.EncTGSRepPart
	if _, err := encTGSRep.Unmarshal(encPlain); err != nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: parse U2U EncTGSRepPart: %w", err)
	}
	if encTGSRep.Nonce != nonce {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: U2U nonce mismatch: got %d, want %d", encTGSRep.Nonce, nonce)
	}

	return tgsRep.Ticket, tgsRep.TicketRaw, encTGSRep.Key.KeyValue, nil
}
