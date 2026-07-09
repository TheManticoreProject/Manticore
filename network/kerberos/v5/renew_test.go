package kerberos

import (
	"bytes"
	"encoding/asn1"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// renewReqBody parses a marshaled TGS-REQ down to its [4] KDC-REQ-BODY.
func renewReqBody(t *testing.T, wire []byte) asn1.RawValue {
	t.Helper()
	var app asn1.RawValue
	if _, err := asn1.Unmarshal(wire, &app); err != nil {
		t.Fatal(err)
	}
	if app.Class != asn1.ClassApplication || app.Tag != messages.MsgTypeTGSReq {
		t.Fatalf("outer tag: class=%d tag=%d", app.Class, app.Tag)
	}
	return seqChild(t, app.Bytes, asn1.ClassContextSpecific, 4)
}

// TestRenewRequestShape verifies a RENEW TGS-REQ sets the renew option (bit 30)
// plus renewable (bit 8), targets krbtgt/REALM, and presents the client's TGT via
// a PA-TGS-REQ AP-REQ (RFC 4120 §3.3.3).
func TestRenewRequestShape(t *testing.T) {
	c := fakeTGTClient(t)

	req, err := c.buildRenewalTGSReq(kdcOptionRenew, 7777)
	if err != nil {
		t.Fatalf("buildRenewalTGSReq: %v", err)
	}
	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := renewReqBody(t, wire)

	// KDCOptions [0]: renew (bit 30) and renewable (bit 8) set; validate (bit 31)
	// and enc-tkt-in-skey (bit 28) clear.
	optElem := seqChild(t, body.Bytes, asn1.ClassContextSpecific, 0)
	var flags asn1.BitString
	if _, err := asn1.Unmarshal(optElem.Bytes, &flags); err != nil {
		t.Fatalf("parse KDCOptions: %v", err)
	}
	if flags.At(kdcOptionRenew) != 1 {
		t.Errorf("renew (bit 30) not set; bytes=% X", flags.Bytes)
	}
	if flags.At(kdcOptionRenewable) != 1 {
		t.Errorf("renewable (bit 8) not set; bytes=% X", flags.Bytes)
	}
	if flags.At(kdcOptionValidate) != 0 {
		t.Errorf("validate (bit 31) unexpectedly set")
	}
	if flags.At(kdcOptionEncTktInSKey) != 0 {
		t.Errorf("enc-tkt-in-skey (bit 28) unexpectedly set")
	}

	// SName [3] must be krbtgt/REALM.
	assertKrbtgtSName(t, body, c.realm)

	// PA-DATA must carry a single PA-TGS-REQ presenting the client's TGT.
	assertPATGSPresentsTGT(t, wire, c.tgtTicketRaw)
}

// TestValidateRequestShape verifies a VALIDATE TGS-REQ sets the validate option
// (bit 31) rather than renew, and otherwise matches the renewal shape.
func TestValidateRequestShape(t *testing.T) {
	c := fakeTGTClient(t)

	req, err := c.buildRenewalTGSReq(kdcOptionValidate, 8888)
	if err != nil {
		t.Fatalf("buildRenewalTGSReq: %v", err)
	}
	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := renewReqBody(t, wire)

	optElem := seqChild(t, body.Bytes, asn1.ClassContextSpecific, 0)
	var flags asn1.BitString
	if _, err := asn1.Unmarshal(optElem.Bytes, &flags); err != nil {
		t.Fatalf("parse KDCOptions: %v", err)
	}
	if flags.At(kdcOptionValidate) != 1 {
		t.Errorf("validate (bit 31) not set; bytes=% X", flags.Bytes)
	}
	if flags.At(kdcOptionRenew) != 0 {
		t.Errorf("renew (bit 30) unexpectedly set")
	}

	assertKrbtgtSName(t, body, c.realm)
	assertPATGSPresentsTGT(t, wire, c.tgtTicketRaw)
}

// TestRenewValidateRequireTGT verifies both operations refuse to run without a
// TGT.
func TestRenewValidateRequireTGT(t *testing.T) {
	c := NewClient("svc", "corp.local", "10.0.0.1").WithPassword("x")
	if err := c.Renew(); err == nil {
		t.Error("expected error renewing without a TGT")
	}
	if err := c.Validate(); err == nil {
		t.Error("expected error validating without a TGT")
	}
}

// assertKrbtgtSName checks the KDC-REQ-BODY's SName [3] is krbtgt/REALM.
func assertKrbtgtSName(t *testing.T, body asn1.RawValue, realm string) {
	t.Helper()
	snameElem := seqChild(t, body.Bytes, asn1.ClassContextSpecific, 3)
	var sname messages.PrincipalName
	if _, err := asn1.Unmarshal(snameElem.Bytes, &sname); err != nil {
		t.Fatalf("parse SName: %v", err)
	}
	if len(sname.NameString) != 2 || sname.NameString[0] != "krbtgt" || sname.NameString[1] != realm {
		t.Errorf("SName not krbtgt/%s: %+v", realm, sname.NameString)
	}
}

// assertPATGSPresentsTGT checks the TGS-REQ [3] PA-DATA holds exactly one
// PA-TGS-REQ whose AP-REQ carries the client's TGT verbatim.
func assertPATGSPresentsTGT(t *testing.T, wire, tgtRaw []byte) {
	t.Helper()
	var app asn1.RawValue
	if _, err := asn1.Unmarshal(wire, &app); err != nil {
		t.Fatal(err)
	}
	paElem := seqChild(t, app.Bytes, asn1.ClassContextSpecific, 3)
	var paList []messages.PAData
	if _, err := asn1.Unmarshal(paElem.Bytes, &paList); err != nil {
		t.Fatalf("parse PA-DATA: %v", err)
	}
	if len(paList) != 1 || paList[0].PADataType != messages.PATGSReq {
		t.Fatalf("expected a single PA-TGS-REQ, got %+v", paList)
	}
	// The AP-REQ (APPLICATION[14]) must embed the TGT (APPLICATION[1]) verbatim
	// under its ticket [3] field, so the raw ticket bytes appear in the padata.
	var apReq asn1.RawValue
	if _, err := asn1.Unmarshal(paList[0].PADataValue, &apReq); err != nil {
		t.Fatalf("parse AP-REQ: %v", err)
	}
	if apReq.Class != asn1.ClassApplication || apReq.Tag != messages.MsgTypeAPReq {
		t.Fatalf("PA-TGS-REQ value is not an AP-REQ: class=%d tag=%d", apReq.Class, apReq.Tag)
	}
	if !bytes.Contains(paList[0].PADataValue, tgtRaw) {
		t.Errorf("PA-TGS-REQ does not present the client's TGT verbatim")
	}
}
