package kerberos

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// TestKerberoastPreloadedTicket exercises Kerberoast end-to-end without a KDC by
// preloading a service ticket: GetTGS returns it directly, and Kerberoast must
// surface the ticket enc-part (the offline-crackable material) with its etype.
func TestKerberoastPreloadedTicket(t *testing.T) {
	cipher := bytes.Repeat([]byte{0xC0}, 48)
	svcTkt := messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   "CORP.LOCAL",
		SName:   messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "dc01.corp.local"}},
		EncPart: messages.EncryptedData{EType: messages.ETypeRC4HMAC, Cipher: cipher},
	}
	raw, err := svcTkt.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	c := NewClient("alice", "corp.local", "10.0.0.1").WithPassword("x")
	st := &ServiceTicket{
		Ticket:       svcTkt,
		TicketRaw:    raw,
		SessionKey:   bytes.Repeat([]byte{0x11}, 16),
		SessionEType: messages.ETypeRC4HMAC,
		Client:       messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{"alice"}},
		CRealm:       "CORP.LOCAL",
		SName:        svcTkt.SName,
		SRealm:       "CORP.LOCAL",
	}
	if err := c.LoadServiceTicket(st); err != nil {
		t.Fatalf("LoadServiceTicket: %v", err)
	}

	res, err := c.Kerberoast("cifs/dc01.corp.local")
	if err != nil {
		t.Fatalf("Kerberoast: %v", err)
	}
	if res.SPN != "cifs/dc01.corp.local" || res.Realm != "CORP.LOCAL" {
		t.Errorf("result identity = %+v", res)
	}
	if res.EType != messages.ETypeRC4HMAC {
		t.Errorf("etype = %d, want RC4", res.EType)
	}
	if !bytes.Equal(res.Cipher, cipher) {
		t.Errorf("cipher not surfaced: got %X, want %X", res.Cipher, cipher)
	}
}

// TestKerberoastRequiresTGT confirms Kerberoast fails cleanly with neither a TGT
// nor a preloaded ticket for the SPN.
func TestKerberoastRequiresTGT(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1").WithPassword("x")
	if _, err := c.Kerberoast("cifs/dc01.corp.local"); err == nil {
		t.Error("expected error roasting without a TGT")
	}
}

// TestBuildASREPRoastReq checks the AS-REP roasting request shape: no PA-DATA
// (the un-authenticated request is what makes the account roastable), the target
// as cname, krbtgt/REALM as sname, and the strongest-first etype list.
func TestBuildASREPRoastReq(t *testing.T) {
	req, err := buildASREPRoastReq("victim", "CORP.LOCAL")
	if err != nil {
		t.Fatalf("buildASREPRoastReq: %v", err)
	}

	if len(req.PAData) != 0 {
		t.Errorf("AS-REP roast request must carry no PA-DATA, got %d", len(req.PAData))
	}
	if req.ReqBody.CName.NameType != messages.NameTypePrincipal ||
		req.ReqBody.CName.NameString[0] != "victim" {
		t.Errorf("cname = %+v, want victim", req.ReqBody.CName)
	}
	if req.ReqBody.SName.NameString[0] != "krbtgt" || req.ReqBody.SName.NameString[1] != "CORP.LOCAL" {
		t.Errorf("sname = %+v, want krbtgt/CORP.LOCAL", req.ReqBody.SName)
	}
	if len(req.ReqBody.EType) == 0 || req.ReqBody.EType[0] != messages.ETypeAES256CTSHMACSHA196 {
		t.Errorf("etype list = %v, want AES256 first", req.ReqBody.EType)
	}
	if req.ReqBody.Nonce <= 0 {
		t.Errorf("nonce = %d, want a positive value", req.ReqBody.Nonce)
	}

	// The request must marshal and round-trip.
	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got messages.ASReq
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(got.PAData) != 0 || got.ReqBody.CName.NameString[0] != "victim" {
		t.Errorf("round-trip changed the request: %+v", got.ReqBody)
	}
}
