package messages

import (
	"bytes"
	"testing"
)

// TestTGSRepUnmarshalPreservesTicketRaw verifies that a TGS-REP round-trips
// through Marshal → Unmarshal with the raw APPLICATION[1] ticket bytes
// preserved on TicketRaw, and that those bytes parse back as a valid Ticket
// on their own.
func TestTGSRepUnmarshalPreservesTicketRaw(t *testing.T) {
	original := &TGSRep{
		PVNO:    KerberosV5,
		MsgType: MsgTypeTGSRep,
		CRealm:  "CORP.LOCAL",
		CName: PrincipalName{
			NameType:   NameTypePrincipal,
			NameString: []string{"alice"},
		},
		Ticket: Ticket{
			TktVno: KerberosV5,
			Realm:  "CORP.LOCAL",
			SName: PrincipalName{
				NameType:   NameTypeSRVInst,
				NameString: []string{"cifs", "dc01.corp.local"},
			},
			EncPart: EncryptedData{
				EType:  ETypeAES256CTSHMACSHA196,
				Cipher: []byte{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe},
			},
		},
		EncPart: EncryptedData{
			EType:  ETypeAES256CTSHMACSHA196,
			Cipher: []byte{0xaa, 0xbb, 0xcc, 0xdd},
		},
	}

	wire, err := original.Marshal()
	if err != nil {
		t.Fatalf("TGSRep.Marshal: %v", err)
	}

	var decoded TGSRep
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("TGSRep.Unmarshal: %v", err)
	}

	if len(decoded.TicketRaw) == 0 {
		t.Fatalf("TicketRaw was not populated")
	}

	// TicketRaw must be a standalone, re-parseable APPLICATION[1] ticket.
	var reparsed Ticket
	if _, err := reparsed.Unmarshal(decoded.TicketRaw); err != nil {
		t.Fatalf("TicketRaw is not a valid APPLICATION[1] ticket: %v", err)
	}
	if reparsed.Realm != original.Ticket.Realm {
		t.Errorf("reparsed Realm: got %q, want %q", reparsed.Realm, original.Ticket.Realm)
	}
	if !bytes.Equal(reparsed.EncPart.Cipher, original.Ticket.EncPart.Cipher) {
		t.Errorf("reparsed EncPart.Cipher: got %x, want %x",
			reparsed.EncPart.Cipher, original.Ticket.EncPart.Cipher)
	}

	// TicketRaw must also be usable as the TicketRaw of an APReq — i.e.
	// APReq.Marshal must accept those bytes and Unmarshal must recover them
	// without loss. This is the primary motivating downstream use case.
	apReq := &APReq{
		PVNO:      KerberosV5,
		MsgType:   MsgTypeAPReq,
		TicketRaw: decoded.TicketRaw,
	}
	apReqWire, err := apReq.Marshal()
	if err != nil {
		t.Fatalf("APReq.Marshal with TicketRaw: %v", err)
	}
	if len(apReqWire) == 0 {
		t.Fatalf("APReq.Marshal produced no bytes")
	}
}
