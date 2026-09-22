package kerberos

import (
	"fmt"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// buildForwardedProxyTGSReq builds an RFC 4120 FORWARDED or PROXY request.
// These exchanges are distinct from Microsoft S4U2Proxy: they present the
// client's own TGT and bind the resulting ticket to caller-supplied addresses.
func (c *KerberosClient) buildForwardedProxyTGSReq(option int, sname messages.PrincipalName, addresses []messages.HostAddress, nonce int) (*messages.TGSReq, error) {
	if option != kdcOptionForwarded && option != kdcOptionProxy {
		return nil, fmt.Errorf("kerberos: unsupported derived-ticket option %d", option)
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("kerberos: forwarded and proxy requests require at least one address")
	}
	body := messages.KDCReqBody{
		KDCOptions: encodeKDCOptions(option), Realm: c.realm, SName: sname,
		Till: c.now().Add(24 * time.Hour), Nonce: nonce,
		EType: c.serviceTicketETypes(), Addresses: addresses,
	}
	apReq, err := c.buildAPReq(body)
	if err != nil {
		return nil, fmt.Errorf("kerberos: build AP-REQ: %w", err)
	}
	return &messages.TGSReq{
		PVNO: messages.KerberosV5, MsgType: messages.MsgTypeTGSReq,
		PAData:  []messages.PAData{{PADataType: messages.PATGSReq, PADataValue: apReq}},
		ReqBody: body,
	}, nil
}

// GetForwardedTGT requests a FORWARDED TGT restricted to addresses.
func (c *KerberosClient) GetForwardedTGT(addresses []messages.HostAddress) (messages.Ticket, []byte, []byte, int, error) {
	sname := messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", c.realm}}
	return c.getForwardedProxyTicket(kdcOptionForwarded, messages.TicketFlagForwarded, sname, addresses, "forwarded")
}

// GetProxyTicket requests an RFC 4120 PROXY service ticket restricted to
// addresses. It does not perform the MS-SFU S4U2Proxy exchange.
func (c *KerberosClient) GetProxyTicket(targetSPN string, addresses []messages.HostAddress) (messages.Ticket, []byte, []byte, int, error) {
	sname, err := parseSPN(targetSPN, c.realm)
	if err != nil {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: parse proxy target SPN %q: %w", targetSPN, err)
	}
	return c.getForwardedProxyTicket(kdcOptionProxy, messages.TicketFlagProxy, sname, addresses, "proxy")
}

func (c *KerberosClient) getForwardedProxyTicket(option, requiredFlag int, sname messages.PrincipalName, addresses []messages.HostAddress, label string) (messages.Ticket, []byte, []byte, int, error) {
	if !c.hasTGT {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: no TGT: call GetTGT first")
	}
	nonce := randomNonce()
	req, err := c.buildForwardedProxyTGSReq(option, sname, addresses, nonce)
	if err != nil {
		return messages.Ticket{}, nil, nil, 0, err
	}
	wire, err := req.Marshal()
	if err != nil {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: marshal %s TGS-REQ: %w", label, err)
	}
	resp, err := c.sendToRealm(c.realm, wire)
	if err != nil {
		return messages.Ticket{}, nil, nil, 0, err
	}
	var krbErr messages.KRBError
	if _, parseErr := krbErr.Unmarshal(resp); parseErr == nil {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: %s error %d: %s", label, krbErr.ErrorCode, krbErr.EText)
	}
	var rep messages.TGSRep
	if _, err := rep.Unmarshal(resp); err != nil {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: parse %s TGS-REP: %w", label, err)
	}
	plain, err := kerbcrypto.Decrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageTGSRepEncSessionKey, rep.EncPart.Cipher)
	if err != nil {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: decrypt %s TGS-REP: %w", label, err)
	}
	var enc messages.EncTGSRepPart
	if _, err := enc.Unmarshal(plain); err != nil {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: parse %s EncTGSRepPart: %w", label, err)
	}
	if enc.Nonce != nonce {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: %s nonce mismatch: got %d, want %d", label, enc.Nonce, nonce)
	}
	if enc.Flags.At(requiredFlag) == 0 {
		return messages.Ticket{}, nil, nil, 0, fmt.Errorf("kerberos: %s reply omitted required ticket flag %d", label, requiredFlag)
	}
	return rep.Ticket, rep.TicketRaw, enc.Key.KeyValue, enc.Key.KeyType, nil
}
