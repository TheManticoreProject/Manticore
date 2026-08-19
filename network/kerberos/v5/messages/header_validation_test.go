package messages

import (
	"bytes"
	"testing"
	"time"
)

type messageMarshaler interface {
	Marshal() ([]byte, error)
}

func marshalHeaderTestMessage(t *testing.T, m messageMarshaler) []byte {
	t.Helper()
	wire, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return wire
}

func replaceExplicitInteger(t *testing.T, wire []byte, tag, old, replacement byte) []byte {
	t.Helper()
	out := append([]byte(nil), wire...)
	pattern := []byte{tag, 0x03, 0x02, 0x01, old}
	i := bytes.Index(out, pattern)
	if i < 0 {
		t.Fatalf("explicit integer %x=%d not found in %x", tag, old, wire)
	}
	out[i+len(pattern)-1] = replacement
	return out
}

func TestUnmarshalRejectsInvalidKerberosHeaders(t *testing.T) {
	now := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	principal := PrincipalName{NameType: NameTypePrincipal, NameString: []string{"alice"}}
	service := PrincipalName{NameType: NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}}
	enc := EncryptedData{EType: ETypeAES256CTSHMACSHA196, Cipher: []byte{1, 2, 3}}
	ticket := Ticket{TktVno: KerberosV5, Realm: "CORP.LOCAL", SName: service, EncPart: enc}
	body := KDCReqBody{
		KDCOptions: NewKerberosFlags(),
		CName:      principal,
		Realm:      "CORP.LOCAL",
		SName:      service,
		Till:       now.Add(time.Hour),
		Nonce:      1,
		EType:      []int{ETypeAES256CTSHMACSHA196},
	}

	tests := []struct {
		name    string
		wire    []byte
		pvnoTag byte
		msgTag  byte
		msgType byte
		decode  func([]byte) error
	}{
		{"ASReq", marshalHeaderTestMessage(t, &ASReq{ReqBody: body}), 0xa1, 0xa2, MsgTypeASReq, func(b []byte) error { var v ASReq; _, err := v.Unmarshal(b); return err }},
		{"TGSReq", marshalHeaderTestMessage(t, &TGSReq{ReqBody: body}), 0xa1, 0xa2, MsgTypeTGSReq, func(b []byte) error { var v TGSReq; _, err := v.Unmarshal(b); return err }},
		{"ASRep", marshalHeaderTestMessage(t, &ASRep{CRealm: "CORP.LOCAL", CName: principal, Ticket: ticket, EncPart: enc}), 0xa0, 0xa1, MsgTypeASRep, func(b []byte) error { var v ASRep; _, err := v.Unmarshal(b); return err }},
		{"TGSRep", marshalHeaderTestMessage(t, &TGSRep{CRealm: "CORP.LOCAL", CName: principal, Ticket: ticket, EncPart: enc}), 0xa0, 0xa1, MsgTypeTGSRep, func(b []byte) error { var v TGSRep; _, err := v.Unmarshal(b); return err }},
		{"APReq", marshalHeaderTestMessage(t, &APReq{Ticket: ticket, Authenticator: enc}), 0xa0, 0xa1, MsgTypeAPReq, func(b []byte) error { var v APReq; _, err := v.Unmarshal(b); return err }},
		{"APRep", marshalHeaderTestMessage(t, &APRep{EncPart: enc}), 0xa0, 0xa1, MsgTypeAPRep, func(b []byte) error { var v APRep; _, err := v.Unmarshal(b); return err }},
		{"KRBError", marshalHeaderTestMessage(t, &KRBError{STime: now, Realm: "CORP.LOCAL", SName: service}), 0xa0, 0xa1, MsgTypeError, func(b []byte) error { var v KRBError; _, err := v.Unmarshal(b); return err }},
		{"KRBCred", marshalHeaderTestMessage(t, &KRBCred{Tickets: []Ticket{ticket}, EncPart: enc}), 0xa0, 0xa1, MsgTypeKRBCred, func(b []byte) error { var v KRBCred; _, err := v.Unmarshal(b); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name+"/version", func(t *testing.T) {
			if err := tc.decode(replaceExplicitInteger(t, tc.wire, tc.pvnoTag, KerberosV5, 4)); err == nil {
				t.Fatal("accepted an invalid protocol version")
			}
		})
		t.Run(tc.name+"/message-type", func(t *testing.T) {
			if err := tc.decode(replaceExplicitInteger(t, tc.wire, tc.msgTag, tc.msgType, 1)); err == nil {
				t.Fatal("accepted an inconsistent message type")
			}
		})
	}

	t.Run("Ticket/version", func(t *testing.T) {
		wire := marshalHeaderTestMessage(t, &ticket)
		var decoded Ticket
		if _, err := decoded.Unmarshal(replaceExplicitInteger(t, wire, 0xa0, KerberosV5, 4)); err == nil {
			t.Fatal("accepted an invalid ticket version")
		}
	})

	t.Run("Authenticator/version", func(t *testing.T) {
		wire := marshalHeaderTestMessage(t, &Authenticator{
			CRealm: "CORP.LOCAL",
			CName:  principal,
			CTime:  now,
		})
		var decoded Authenticator
		if _, err := decoded.Unmarshal(replaceExplicitInteger(t, wire, 0xa0, KerberosV5, 4)); err == nil {
			t.Fatal("accepted an invalid authenticator version")
		}
	})
}
