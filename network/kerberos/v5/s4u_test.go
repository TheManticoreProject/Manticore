package kerberos

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/sfu"
)

// paByType returns the first PA-DATA element of the given type, or nil.
func paByType(pas []messages.PAData, typ int) *messages.PAData {
	for i := range pas {
		if pas[i].PADataType == typ {
			return &pas[i]
		}
	}
	return nil
}

// TestS4U2SelfRequestShape checks the S4U2Self TGS-REQ builder: it carries a
// PA-TGS-REQ, a PA-PAC-REQUEST and a PA-FOR-USER naming the impersonated user,
// requests the service's own account as sname, and sets canonicalize.
func TestS4U2SelfRequestShape(t *testing.T) {
	c := fakeTGTClient(t) // client "alice" in CORP.LOCAL with a populated TGT

	req, err := c.buildS4U2SelfTGSReq("victim", "", 4242)
	if err != nil {
		t.Fatalf("buildS4U2SelfTGSReq: %v", err)
	}

	if paByType(req.PAData, messages.PATGSReq) == nil {
		t.Error("missing PA-TGS-REQ")
	}
	if paByType(req.PAData, messages.PAPACRequest) == nil {
		t.Error("missing PA-PAC-REQUEST")
	}
	forUser := paByType(req.PAData, iana.PAForUser)
	if forUser == nil {
		t.Fatal("missing PA-FOR-USER")
	}
	parsed, err := sfu.ParsePAForUser(forUser.PADataValue)
	if err != nil {
		t.Fatalf("ParsePAForUser: %v", err)
	}
	if len(parsed.UserName.NameString) != 1 || parsed.UserName.NameString[0] != "victim" {
		t.Errorf("PA-FOR-USER user = %+v, want victim", parsed.UserName)
	}
	if parsed.UserRealm != "CORP.LOCAL" {
		t.Errorf("PA-FOR-USER realm = %q, want CORP.LOCAL (client realm default)", parsed.UserRealm)
	}
	// The PA-FOR-USER checksum must verify under the TGT session key.
	if !sfu.VerifyPAForUser(parsed, c.sessionKey, c.sessionEType) {
		t.Error("PA-FOR-USER checksum does not verify under the TGT session key")
	}

	// sname is the service's own account (the client principal).
	if len(req.ReqBody.SName.NameString) != 1 || req.ReqBody.SName.NameString[0] != "alice" {
		t.Errorf("sname = %+v, want [alice]", req.ReqBody.SName)
	}
	if req.ReqBody.KDCOptions.At(kdcOptionCanonicalize) != 1 {
		t.Error("canonicalize option not set")
	}

	// The request must marshal cleanly.
	if _, err := req.Marshal(); err != nil {
		t.Fatalf("Marshal: %v", err)
	}
}

// TestS4U2SelfImpersonateRealm confirms an explicit impersonation realm is
// uppercased and carried in PA-FOR-USER.
func TestS4U2SelfImpersonateRealm(t *testing.T) {
	c := fakeTGTClient(t)
	req, err := c.buildS4U2SelfTGSReq("victim", "other.realm", 1)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := sfu.ParsePAForUser(paByType(req.PAData, iana.PAForUser).PADataValue)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.UserRealm != "OTHER.REALM" {
		t.Errorf("PA-FOR-USER realm = %q, want OTHER.REALM", parsed.UserRealm)
	}
}

// TestS4U2ProxyRequestShape checks the S4U2Proxy TGS-REQ builder: it carries a
// PA-TGS-REQ and PA-PAC-OPTIONS, sets the cname-in-addl-tkt option, targets the
// requested SPN, and carries the S4U2Self ticket verbatim in additional-tickets.
func TestS4U2ProxyRequestShape(t *testing.T) {
	c := fakeTGTClient(t)

	selfTkt := messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   "CORP.LOCAL",
		SName:   messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{"alice"}},
		EncPart: messages.EncryptedData{EType: messages.ETypeAES256CTSHMACSHA196, Cipher: bytes.Repeat([]byte{0x9a}, 16)},
	}
	selfRaw, err := selfTkt.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	sname, err := parseSPN("cifs/target.corp.local", c.realm)
	if err != nil {
		t.Fatal(err)
	}
	req, err := c.buildS4U2ProxyTGSReq(sname, selfRaw, 7)
	if err != nil {
		t.Fatalf("buildS4U2ProxyTGSReq: %v", err)
	}

	if paByType(req.PAData, messages.PATGSReq) == nil {
		t.Error("missing PA-TGS-REQ")
	}
	if paByType(req.PAData, iana.PAPACOptions) == nil {
		t.Error("missing PA-PAC-OPTIONS")
	}
	if req.ReqBody.KDCOptions.At(kdcOptionCNameInAddlTkt) != 1 {
		t.Error("cname-in-addl-tkt option not set")
	}
	if req.ReqBody.SName.NameString[0] != "cifs" || req.ReqBody.SName.NameString[1] != "target.corp.local" {
		t.Errorf("sname = %+v, want cifs/target.corp.local", req.ReqBody.SName)
	}
	if len(req.ReqBody.AdditTicketsRaw) != 1 || !bytes.Equal(req.ReqBody.AdditTicketsRaw[0], selfRaw) {
		t.Error("S4U2Self ticket not carried verbatim in additional-tickets")
	}

	// Round-trip through the wire: additional-tickets must survive.
	wire, err := req.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var got messages.TGSReq
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("TGSReq.Unmarshal: %v", err)
	}
	if len(got.ReqBody.AdditTicketsRaw) != 1 || !bytes.Equal(got.ReqBody.AdditTicketsRaw[0], selfRaw) {
		t.Error("additional ticket not preserved through the wire")
	}
}

// TestS4UInputValidation covers the guard clauses on the exported S4U methods.
func TestS4UInputValidation(t *testing.T) {
	noTGT := NewClient("svc", "corp.local", "10.0.0.1").WithPassword("x")
	if _, _, _, err := noTGT.S4U2Self("victim", ""); err == nil {
		t.Error("S4U2Self without a TGT should error")
	}
	if _, _, _, err := noTGT.S4U2Proxy("cifs/target", []byte{1}); err == nil {
		t.Error("S4U2Proxy without a TGT should error")
	}

	c := fakeTGTClient(t)
	if _, _, _, err := c.S4U2Self("", ""); err == nil {
		t.Error("S4U2Self with an empty impersonation user should error")
	}
	if _, _, _, err := c.S4U2Proxy("cifs/target", nil); err == nil {
		t.Error("S4U2Proxy without the S4U2Self ticket should error")
	}
}
