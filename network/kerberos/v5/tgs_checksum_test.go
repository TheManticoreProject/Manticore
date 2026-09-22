package kerberos

import (
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

func assertTGSBodyChecksum(t *testing.T, c *KerberosClient, req *messages.TGSReq) {
	t.Helper()
	var paValue []byte
	for _, pa := range req.PAData {
		if pa.PADataType == messages.PATGSReq {
			paValue = pa.PADataValue
			break
		}
	}
	if len(paValue) == 0 {
		t.Fatal("TGS-REQ has no PA-TGS-REQ")
	}
	var apReq messages.APReq
	if _, err := apReq.Unmarshal(paValue); err != nil {
		t.Fatalf("parse PA-TGS-REQ AP-REQ: %v", err)
	}
	plain, err := kerbcrypto.Decrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageTGSReqPAAPReqAuthen, apReq.Authenticator.Cipher)
	if err != nil {
		t.Fatalf("decrypt PA-TGS-REQ authenticator: %v", err)
	}
	var auth messages.Authenticator
	if _, err := auth.Unmarshal(plain); err != nil {
		t.Fatalf("parse PA-TGS-REQ authenticator: %v", err)
	}
	if auth.Cksum == nil {
		t.Fatal("PA-TGS-REQ authenticator has no checksum")
	}
	body, err := messages.EncodeKDCReqBody(req.ReqBody)
	if err != nil {
		t.Fatalf("encode KDC-REQ-BODY: %v", err)
	}
	if !kerbcrypto.VerifyChecksum(auth.Cksum.CKSumType, c.sessionKey, kerbcrypto.KeyUsageTGSReqAuthCksum, body, auth.Cksum.Checksum) {
		t.Fatal("PA-TGS-REQ authenticator checksum does not cover KDC-REQ-BODY")
	}

	tampered := req.ReqBody
	tampered.Nonce++
	tamperedBody, err := messages.EncodeKDCReqBody(tampered)
	if err != nil {
		t.Fatalf("encode tampered KDC-REQ-BODY: %v", err)
	}
	if kerbcrypto.VerifyChecksum(auth.Cksum.CKSumType, c.sessionKey, kerbcrypto.KeyUsageTGSReqAuthCksum, tamperedBody, auth.Cksum.Checksum) {
		t.Fatal("PA-TGS-REQ authenticator checksum accepted a modified body")
	}
}

func TestTGSRequestBuildersChecksumRequestBody(t *testing.T) {
	c := fakeTGTClient(t)
	service := messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "server.corp.local"}}

	normalBody := messages.KDCReqBody{
		KDCOptions: kdcOptionsForTGSReq(), Realm: c.realm, SName: service,
		Till: c.now().Add(24 * 60 * 60 * 1e9), Nonce: 100, EType: c.serviceTicketETypes(),
	}
	normalAPReq, err := c.buildAPReq(normalBody)
	if err != nil {
		t.Fatalf("build normal AP-REQ: %v", err)
	}
	assertTGSBodyChecksum(t, c, &messages.TGSReq{
		PAData:  []messages.PAData{{PADataType: messages.PATGSReq, PADataValue: normalAPReq}},
		ReqBody: normalBody,
	})

	builders := []struct {
		name string
		fn   func() (*messages.TGSReq, error)
	}{
		{"renew", func() (*messages.TGSReq, error) { return c.buildRenewalTGSReq(kdcOptionRenew, 101) }},
		{"s4u2self", func() (*messages.TGSReq, error) { return c.buildS4U2SelfTGSReq("victim", c.realm, 102) }},
		{"s4u2proxy", func() (*messages.TGSReq, error) { return c.buildS4U2ProxyTGSReq(service, c.tgtTicketRaw, 103) }},
		{"u2u", func() (*messages.TGSReq, error) { return c.buildU2UTGSReq("victim", c.realm, c.tgtTicketRaw, 104) }},
		{"sapphire", func() (*messages.TGSReq, error) { return c.buildSapphireTGSReq("victim", c.realm, 105) }},
	}
	for _, tc := range builders {
		t.Run(tc.name, func(t *testing.T) {
			req, err := tc.fn()
			if err != nil {
				t.Fatalf("build request: %v", err)
			}
			assertTGSBodyChecksum(t, c, req)
		})
	}
}

func TestFASTTGSAuthenticatorChecksumsInnerBody(t *testing.T) {
	c := fakeTGTClient(t)
	body := c.tgsReqBody(c.realm, messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"ldap", "dc.corp.local"}}, 200)
	subkey := make([]byte, kerbcrypto.KeyLen(c.sessionEType))
	apReq, err := c.buildTGSAPReqWithSubkey(body, c.tgtTicket, c.tgtTicketRaw, c.sessionKey, c.sessionEType, subkey)
	if err != nil {
		t.Fatalf("build FAST TGS AP-REQ: %v", err)
	}
	assertTGSBodyChecksum(t, c, &messages.TGSReq{
		PAData:  []messages.PAData{{PADataType: messages.PATGSReq, PADataValue: apReq}},
		ReqBody: body,
	})
}
