package messages

import (
	"bytes"
	"testing"
	"time"
)

// tstTime is a fixed UTC, second-truncated timestamp used across the round-trip
// tests (KerberosTime carries no sub-second precision).
var tstTime = time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)

func cliName() PrincipalName {
	return PrincipalName{NameType: NameTypePrincipal, NameString: []string{"alice"}}
}

func tgtSName() PrincipalName {
	return PrincipalName{NameType: NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}}
}

// TestASReqRoundTrip marshals an AS-REQ (with PA-DATA and a full request body)
// and unmarshals it, checking the wire form preserves every field.
func TestASReqRoundTrip(t *testing.T) {
	orig := &ASReq{
		PVNO:    KerberosV5,
		MsgType: MsgTypeASReq,
		PAData: []PAData{
			{PADataType: PAEncTimestamp, PADataValue: []byte{0x01, 0x02, 0x03}},
			{PADataType: PAPACRequest, PADataValue: []byte{0x30, 0x05, 0xa0, 0x03, 0x01, 0x01, 0xff}},
		},
		ReqBody: KDCReqBody{
			KDCOptions: NewKerberosFlags(1, 3, 8),
			CName:      cliName(),
			Realm:      "CORP.LOCAL",
			SName:      tgtSName(),
			Till:       tstTime,
			Nonce:      0x2A2A2A2A,
			EType:      []int{ETypeAES256CTSHMACSHA196, ETypeAES128CTSHMACSHA196, ETypeRC4HMAC},
		},
	}
	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got ASReq
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.MsgType != MsgTypeASReq || got.PVNO != KerberosV5 {
		t.Errorf("header: pvno=%d msgtype=%d", got.PVNO, got.MsgType)
	}
	if len(got.PAData) != 2 || got.PAData[0].PADataType != PAEncTimestamp ||
		!bytes.Equal(got.PAData[1].PADataValue, orig.PAData[1].PADataValue) {
		t.Errorf("PAData not preserved: %+v", got.PAData)
	}
	if got.ReqBody.Realm != "CORP.LOCAL" || got.ReqBody.CName.NameString[0] != "alice" {
		t.Errorf("names not preserved: %+v", got.ReqBody)
	}
	if got.ReqBody.SName.NameString[1] != "CORP.LOCAL" {
		t.Errorf("sname not preserved: %+v", got.ReqBody.SName)
	}
	if got.ReqBody.Nonce != orig.ReqBody.Nonce {
		t.Errorf("nonce: got %d want %d", got.ReqBody.Nonce, orig.ReqBody.Nonce)
	}
	if !got.ReqBody.Till.Equal(tstTime) {
		t.Errorf("till: got %v want %v", got.ReqBody.Till, tstTime)
	}
	if len(got.ReqBody.EType) != 3 || got.ReqBody.EType[0] != ETypeAES256CTSHMACSHA196 {
		t.Errorf("etype list not preserved: %v", got.ReqBody.EType)
	}
}

// TestTGSReqRoundTrip marshals a TGS-REQ carrying a PA-TGS-REQ and an
// additional-ticket (verbatim APPLICATION[1] bytes) and confirms the body and
// additional-ticket survive the round-trip.
func TestTGSReqRoundTrip(t *testing.T) {
	addl := Ticket{
		TktVno:  KerberosV5,
		Realm:   "CORP.LOCAL",
		SName:   tgtSName(),
		EncPart: EncryptedData{EType: ETypeAES256CTSHMACSHA196, Cipher: bytes.Repeat([]byte{0x5a}, 16)},
	}
	addlRaw, err := addl.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	orig := &TGSReq{
		PVNO:    KerberosV5,
		MsgType: MsgTypeTGSReq,
		PAData:  []PAData{{PADataType: PATGSReq, PADataValue: []byte{0xaa, 0xbb}}},
		ReqBody: KDCReqBody{
			KDCOptions:      NewKerberosFlags(1, 8, 15),
			Realm:           "CORP.LOCAL",
			SName:           PrincipalName{NameType: NameTypeSRVInst, NameString: []string{"cifs", "dc01.corp.local"}},
			Till:            tstTime,
			Nonce:           4242,
			EType:           []int{ETypeAES256CTSHMACSHA196, ETypeRC4HMAC},
			AdditTicketsRaw: [][]byte{addlRaw},
		},
	}
	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got TGSReq
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(got.PAData) != 1 || got.PAData[0].PADataType != PATGSReq {
		t.Errorf("PA-TGS-REQ not preserved: %+v", got.PAData)
	}
	if got.ReqBody.SName.NameString[0] != "cifs" || got.ReqBody.Nonce != 4242 {
		t.Errorf("body not preserved: %+v", got.ReqBody)
	}
	if len(got.ReqBody.AdditTicketsRaw) != 1 {
		t.Fatalf("additional-tickets not decoded: %d", len(got.ReqBody.AdditTicketsRaw))
	}
	if !bytes.Equal(got.ReqBody.AdditTicketsRaw[0], addlRaw) {
		t.Errorf("additional ticket not preserved verbatim:\n got  %X\n want %X",
			got.ReqBody.AdditTicketsRaw[0], addlRaw)
	}
	// The recovered raw ticket must re-parse as a standalone APPLICATION[1] ticket.
	var reAddl Ticket
	if _, err := reAddl.Unmarshal(got.ReqBody.AdditTicketsRaw[0]); err != nil {
		t.Errorf("recovered additional ticket does not parse: %v", err)
	}
}

// TestASRepRoundTrip marshals an AS-REP and confirms the ticket, enc-part and
// principal names round-trip, and that TicketRaw is a standalone ticket.
func TestASRepRoundTrip(t *testing.T) {
	orig := &ASRep{
		PVNO:    KerberosV5,
		MsgType: MsgTypeASRep,
		CRealm:  "CORP.LOCAL",
		CName:   cliName(),
		Ticket: Ticket{
			TktVno:  KerberosV5,
			Realm:   "CORP.LOCAL",
			SName:   tgtSName(),
			EncPart: EncryptedData{EType: ETypeAES256CTSHMACSHA196, Cipher: []byte{0xde, 0xad, 0xbe, 0xef}},
		},
		EncPart: EncryptedData{EType: ETypeAES256CTSHMACSHA196, Cipher: bytes.Repeat([]byte{0x11}, 40)},
	}
	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got ASRep
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.MsgType != MsgTypeASRep || got.CRealm != "CORP.LOCAL" || got.CName.NameString[0] != "alice" {
		t.Errorf("header/names not preserved: %+v", got)
	}
	if got.Ticket.Realm != "CORP.LOCAL" || !bytes.Equal(got.Ticket.EncPart.Cipher, orig.Ticket.EncPart.Cipher) {
		t.Errorf("ticket not preserved: %+v", got.Ticket)
	}
	if !bytes.Equal(got.EncPart.Cipher, orig.EncPart.Cipher) || got.EncPart.EType != ETypeAES256CTSHMACSHA196 {
		t.Errorf("enc-part not preserved: %+v", got.EncPart)
	}
	var reparsed Ticket
	if _, err := reparsed.Unmarshal(got.TicketRaw); err != nil {
		t.Fatalf("TicketRaw not a standalone ticket: %v", err)
	}
}
