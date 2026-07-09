package kerberos

import (
	"fmt"
	"strings"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/sfu"
)

// S4U2Self performs the MS-SFU S4U2Self exchange: the service (this client's
// principal), holding its own TGT, requests a service ticket to itself on behalf
// of the user (impersonateUser, impersonateRealm), identified only by name. It
// is the first half of constrained delegation and, on its own, a way to obtain a
// usable service ticket (with the target user's PAC) for any user without their
// secret — subject to the account's delegation configuration.
//
// GetTGT must have succeeded first (the client must hold its service TGT). If
// impersonateRealm is empty the client's realm is used. Returns the service
// ticket, its raw APPLICATION[1] bytes, and the ticket session key.
func (c *KerberosClient) S4U2Self(impersonateUser, impersonateRealm string) (messages.Ticket, []byte, []byte, error) {
	if !c.hasTGT {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: no TGT: call GetTGT first")
	}
	if impersonateUser == "" {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: S4U2Self requires a user to impersonate")
	}
	if impersonateRealm == "" {
		impersonateRealm = c.realm
	} else {
		impersonateRealm = strings.ToUpper(impersonateRealm)
	}

	// PA-FOR-USER identifies the impersonated user, protected by a checksum keyed
	// with this service's TGT session key.
	userName := messages.PrincipalName{
		NameType:   messages.NameTypePrincipal,
		NameString: []string{impersonateUser},
	}
	paForUser, err := sfu.BuildPAForUser(userName, impersonateRealm, c.sessionKey, c.sessionEType)
	if err != nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: build PA-FOR-USER: %w", err)
	}

	// AP-REQ over the service's own TGT.
	apReqBytes, err := c.buildAPReq()
	if err != nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: build AP-REQ: %w", err)
	}

	// The requested server is the service itself (its own account).
	self := messages.PrincipalName{
		NameType:   messages.NameTypePrincipal,
		NameString: []string{c.username},
	}

	nonce := randomNonce()
	tgsReq := &messages.TGSReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeTGSReq,
		PAData: []messages.PAData{
			{PADataType: messages.PATGSReq, PADataValue: apReqBytes},
			paForUser,
			{PADataType: messages.PAPACRequest, PADataValue: []byte{0x30, 0x05, 0xa0, 0x03, 0x01, 0x01, 0xff}},
		},
		ReqBody: messages.KDCReqBody{
			KDCOptions: kdcOptionsForTGSReq(),
			Realm:      c.realm,
			SName:      self,
			Till:       time.Now().UTC().Add(24 * time.Hour),
			Nonce:      nonce,
			EType: []int{
				messages.ETypeAES256CTSHMACSHA196,
				messages.ETypeAES128CTSHMACSHA196,
				messages.ETypeRC4HMAC,
			},
		},
	}

	tgsReqBytes, err := tgsReq.Marshal()
	if err != nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: marshal S4U2Self TGS-REQ: %w", err)
	}
	resp, err := kdcSend(c.kdcHost, defaultKDCPort, tgsReqBytes)
	if err != nil {
		return messages.Ticket{}, nil, nil, err
	}

	var krbErr messages.KRBError
	if _, parseErr := krbErr.Unmarshal(resp); parseErr == nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: S4U2Self error %d: %s", krbErr.ErrorCode, krbErr.EText)
	}

	var tgsRep messages.TGSRep
	if _, err := tgsRep.Unmarshal(resp); err != nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: parse S4U2Self TGS-REP: %w", err)
	}

	encPlain, err := kerbcrypto.Decrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageTGSRepEncSessionKey, tgsRep.EncPart.Cipher)
	if err != nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: decrypt S4U2Self TGS-REP enc-part: %w", err)
	}
	var encTGSRep messages.EncTGSRepPart
	if _, err := encTGSRep.Unmarshal(encPlain); err != nil {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: parse S4U2Self EncTGSRepPart: %w", err)
	}
	if encTGSRep.Nonce != nonce {
		return messages.Ticket{}, nil, nil, fmt.Errorf("kerberos: S4U2Self nonce mismatch: got %d, want %d", encTGSRep.Nonce, nonce)
	}

	return tgsRep.Ticket, tgsRep.TicketRaw, encTGSRep.Key.KeyValue, nil
}
