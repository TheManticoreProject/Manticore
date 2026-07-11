package kerberos

import (
	"bytes"
	"encoding/asn1"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// asReqBody parses a marshaled AS-REQ down to its [4] KDC-REQ-BODY.
func asReqBody(t *testing.T, wire []byte) asn1.RawValue {
	t.Helper()
	var app asn1.RawValue
	if _, err := asn1.Unmarshal(wire, &app); err != nil {
		t.Fatal(err)
	}
	if app.Class != asn1.ClassApplication || app.Tag != messages.MsgTypeASReq {
		t.Fatalf("outer tag: class=%d tag=%d", app.Class, app.Tag)
	}
	return seqChild(t, app.Bytes, asn1.ClassContextSpecific, 4)
}

// TestAsReqKDCOptionsPostdate verifies the exact wire bits of the postdated
// AS-REQ options: forwardable + proxiable + renewable plus allow-postdate
// (bit 5 -> byte 0 0x04) and postdated (bit 6 -> byte 0 0x02). Without a
// postdate request, the standard AS-REQ options are used.
func TestAsReqKDCOptionsPostdate(t *testing.T) {
	c := NewClient("svc", "corp.local", "10.0.0.1").WithPassword("x")

	// No postdate configured: standard AS-REQ options.
	if got := c.asReqKDCOptions(); !bytes.Equal(got.Bytes, kdcOptionsForASReq().Bytes) {
		t.Errorf("non-postdated options: got %x, want %x", got.Bytes, kdcOptionsForASReq().Bytes)
	}

	// Postdated: forwardable (0x40) + proxiable (0x10) + allow-postdate (0x04)
	// + postdated (0x02) in byte 0 => 0x56; renewable (0x80) in byte 1.
	c.WithPostdate(time.Date(2027, 1, 2, 15, 4, 5, 0, time.UTC))
	want := []byte{0x56, 0x80, 0x00, 0x00}
	got := c.asReqKDCOptions()
	if got.BitLength != 32 {
		t.Errorf("BitLength: got %d, want 32", got.BitLength)
	}
	if !bytes.Equal(got.Bytes, want) {
		t.Errorf("Bytes: got %x, want %x", got.Bytes, want)
	}
}

// TestPostdateASReqShape verifies a postdated AS-REQ sets the allow-postdate
// (bit 5) and postdated (bit 6) options, carries the requested future start
// time in from [4], and derives its endtime till [5] from that start time
// (RFC 4120 §3.3).
func TestPostdateASReqShape(t *testing.T) {
	start := time.Date(2027, 1, 2, 15, 4, 5, 0, time.UTC)
	c := NewClient("svc", "corp.local", "10.0.0.1").WithPassword("x").WithPostdate(start)

	req := c.buildASReq(nil, 4242)
	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := asReqBody(t, wire)

	// KDCOptions [0]: allow-postdate (bit 5) and postdated (bit 6) set; renew
	// (bit 30) and validate (bit 31) clear.
	optElem := seqChild(t, body.Bytes, asn1.ClassContextSpecific, 0)
	var flags asn1.BitString
	if _, err := asn1.Unmarshal(optElem.Bytes, &flags); err != nil {
		t.Fatalf("parse KDCOptions: %v", err)
	}
	if flags.At(kdcOptionAllowPostdate) != 1 {
		t.Errorf("allow-postdate (bit 5) not set; bytes=% X", flags.Bytes)
	}
	if flags.At(kdcOptionPostdated) != 1 {
		t.Errorf("postdated (bit 6) not set; bytes=% X", flags.Bytes)
	}
	if flags.At(kdcOptionRenew) != 0 || flags.At(kdcOptionValidate) != 0 {
		t.Errorf("renew/validate unexpectedly set; bytes=% X", flags.Bytes)
	}

	// from [4] must equal the requested start time.
	fromElem := seqChild(t, body.Bytes, asn1.ClassContextSpecific, 4)
	var from time.Time
	if _, err := asn1.Unmarshal(fromElem.Bytes, &from); err != nil {
		t.Fatalf("parse from: %v", err)
	}
	if !from.Equal(start) {
		t.Errorf("from: got %s, want %s", from, start)
	}

	// till [5] must be the endtime derived from the start time (start + 24h).
	tillElem := seqChild(t, body.Bytes, asn1.ClassContextSpecific, 5)
	var till time.Time
	if _, err := asn1.Unmarshal(tillElem.Bytes, &till); err != nil {
		t.Fatalf("parse till: %v", err)
	}
	if !till.Equal(start.Add(24 * time.Hour)) {
		t.Errorf("till: got %s, want %s", till, start.Add(24*time.Hour))
	}
}

// TestNonPostdatedASReqOmitsFrom verifies that an ordinary (non-postdated)
// AS-REQ omits the optional from [4] field entirely.
func TestNonPostdatedASReqOmitsFrom(t *testing.T) {
	c := NewClient("svc", "corp.local", "10.0.0.1").WithPassword("x")
	wire, err := c.buildASReq(nil, 4242).Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := asReqBody(t, wire)

	// Iterate the body: no [4] element must be present.
	var seq asn1.RawValue
	if _, err := asn1.Unmarshal(body.Bytes, &seq); err == nil {
		rest := seq.Bytes
		for len(rest) > 0 {
			var e asn1.RawValue
			r, err := asn1.Unmarshal(rest, &e)
			if err != nil {
				t.Fatalf("iterate body: %v", err)
			}
			if e.Class == asn1.ClassContextSpecific && e.Tag == 4 {
				t.Fatalf("from [4] must be absent from a non-postdated AS-REQ")
			}
			rest = r
		}
	}
}

// TestPostdateFlagsDecode verifies the ticket-flag accessors decode a crafted
// postdated reply: a TGT carrying MAY-POSTDATE (bit 5), POSTDATED (bit 6) and
// INVALID (bit 7), as a KDC returns for a postdated request (RFC 4120 §3.3.3).
func TestPostdateFlagsDecode(t *testing.T) {
	c := NewClient("svc", "corp.local", "10.0.0.1").WithPassword("x")
	c.tgtEnc = messages.EncASRepPart{
		Flags: messages.NewKerberosFlags(
			messages.TicketFlagMayPostdate,
			messages.TicketFlagPostdated,
			messages.TicketFlagInvalid,
		),
	}

	// may-postdate (0x04) + postdated (0x02) + invalid (0x01) all in byte 0.
	if want := byte(0x07); c.tgtEnc.Flags.Bytes[0] != want {
		t.Errorf("flags byte 0: got %#x, want %#x", c.tgtEnc.Flags.Bytes[0], want)
	}
	if !c.TGTMayPostdate() {
		t.Error("TGTMayPostdate: got false, want true")
	}
	if !c.TGTPostdated() {
		t.Error("TGTPostdated: got false, want true")
	}
	if !c.TGTInvalid() {
		t.Error("TGTInvalid: got false, want true")
	}

	// A plain TGT (no postdating flags) must report all three false.
	c.tgtEnc.Flags = messages.NewKerberosFlags(messages.TicketFlagForwardable, messages.TicketFlagRenewable)
	if c.TGTMayPostdate() || c.TGTPostdated() || c.TGTInvalid() {
		t.Error("plain TGT reported a postdating flag")
	}
}
