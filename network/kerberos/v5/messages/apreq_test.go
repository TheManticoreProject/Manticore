package messages

import (
	"bytes"
	"encoding/asn1"
	"testing"
)

// TestAPReqMarshalUnmarshalRoundtrip verifies that marshaling an APReq and then
// unmarshaling the bytes recovers the same field values, including the ticket
// parsed into r.Ticket and the raw ticket bytes preserved in r.TicketRaw.
// This is a regression test for the double-wrapped ticket bug where
// APReq.Unmarshal re-marshaled inner.Ticket (adding an unwanted [3] EXPLICIT
// layer) before handing the bytes to Ticket.Unmarshal.
func TestAPReqMarshalUnmarshalRoundtrip(t *testing.T) {
	original := &APReq{
		PVNO:    KerberosV5,
		MsgType: MsgTypeAPReq,
		APOptions: asn1.BitString{
			Bytes:     []byte{0x00, 0x00, 0x00, 0x00},
			BitLength: 32,
		},
		Ticket: Ticket{
			TktVno: KerberosV5,
			Realm:  "CORP.LOCAL",
			SName: PrincipalName{
				NameType:   NameTypeSRVInst,
				NameString: []string{"krbtgt", "CORP.LOCAL"},
			},
			EncPart: EncryptedData{
				EType:  ETypeAES256CTSHMACSHA196,
				Cipher: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			},
		},
		Authenticator: EncryptedData{
			EType:  ETypeAES256CTSHMACSHA196,
			Cipher: []byte{0xaa, 0xbb, 0xcc, 0xdd},
		},
	}

	wire, err := original.Marshal()
	if err != nil {
		t.Fatalf("APReq.Marshal: %v", err)
	}

	var decoded APReq
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("APReq.Unmarshal: %v", err)
	}

	if decoded.PVNO != original.PVNO {
		t.Errorf("PVNO: got %d, want %d", decoded.PVNO, original.PVNO)
	}
	if decoded.MsgType != original.MsgType {
		t.Errorf("MsgType: got %d, want %d", decoded.MsgType, original.MsgType)
	}
	if decoded.Ticket.Realm != original.Ticket.Realm {
		t.Errorf("Ticket.Realm: got %q, want %q", decoded.Ticket.Realm, original.Ticket.Realm)
	}
	if decoded.Ticket.TktVno != original.Ticket.TktVno {
		t.Errorf("Ticket.TktVno: got %d, want %d", decoded.Ticket.TktVno, original.Ticket.TktVno)
	}
	if !bytes.Equal(decoded.Ticket.EncPart.Cipher, original.Ticket.EncPart.Cipher) {
		t.Errorf("Ticket.EncPart.Cipher: got %x, want %x",
			decoded.Ticket.EncPart.Cipher, original.Ticket.EncPart.Cipher)
	}
	if !bytes.Equal(decoded.Authenticator.Cipher, original.Authenticator.Cipher) {
		t.Errorf("Authenticator.Cipher: got %x, want %x",
			decoded.Authenticator.Cipher, original.Authenticator.Cipher)
	}

	// TicketRaw must hold the APPLICATION[1] ticket TLV so it can be re-emitted
	// verbatim by a subsequent Marshal.
	if len(decoded.TicketRaw) == 0 {
		t.Fatalf("TicketRaw was not populated")
	}
	var reparsed Ticket
	if _, err := reparsed.Unmarshal(decoded.TicketRaw); err != nil {
		t.Fatalf("TicketRaw is not a valid APPLICATION[1] ticket: %v", err)
	}
	if reparsed.Realm != original.Ticket.Realm {
		t.Errorf("TicketRaw reparse Realm: got %q, want %q", reparsed.Realm, original.Ticket.Realm)
	}
}
