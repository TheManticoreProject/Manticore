package messages

import (
	"encoding/asn1"
	"fmt"
)

// TGSRep is a Kerberos TGS-REP (Ticket Granting Service Reply) message,
// APPLICATION[13], as defined in RFC 4120 Section 5.4.2.
// It is sent by the TGS in response to a successful TGS-REQ.
type TGSRep struct {
	// PVNO is the Kerberos protocol version (always 5).
	PVNO int
	// MsgType is the message type (always MsgTypeTGSRep = 13).
	MsgType int
	// PAData contains pre-authentication data (rarely set in TGS-REP).
	PAData []PAData
	// CRealm is the realm of the client.
	CRealm string
	// CName is the client's principal name.
	CName PrincipalName
	// Ticket is the issued service ticket.
	Ticket Ticket
	// EncPart is the encrypted reply body, decryptable with the TGT session key.
	EncPart EncryptedData
}

// Marshal encodes the TGS-REP as an ASN.1 APPLICATION[13] wrapped SEQUENCE.
func (r *TGSRep) Marshal() ([]byte, error) {
	tkt_bytes, err := r.Ticket.Marshal()
	if err != nil {
		return nil, err
	}
	// Pre-encode [5] EXPLICIT { APPLICATION[1] bytes }.
	// Go ignores explicit,tag:N for asn1.RawValue with FullBytes set, so we build
	// the [5] wrapper manually using Bytes (which Go wraps with Class/Tag/IsCompound).
	tkt_raw := asn1.RawValue{
		Class:      asn1.ClassContextSpecific,
		Tag:        5,
		IsCompound: true,
		Bytes:      tkt_bytes,
	}

	inner := kdcRepInner{
		PVNO:    KerberosV5,
		MsgType: MsgTypeTGSRep,
		PAData:  r.PAData,
		CRealm:  r.CRealm,
		CName:   r.CName,
		Ticket:  tkt_raw,
		EncPart: r.EncPart,
	}
	seq_bytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(MsgTypeTGSRep, seq_bytes)
}

// Unmarshal decodes a TGS-REP from an ASN.1 APPLICATION[13] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (r *TGSRep) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, MsgTypeTGSRep)
	if err != nil {
		return 0, fmt.Errorf("tgsrep: %w", err)
	}

	var inner kdcRepInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("tgsrep inner unmarshal: %w", err)
	}

	r.PVNO = inner.PVNO
	r.MsgType = inner.MsgType
	r.PAData = inner.PAData
	r.CRealm = inner.CRealm
	r.CName = inner.CName
	r.EncPart = inner.EncPart

	// inner.Ticket.Bytes is the APPLICATION[1] ticket (Go does not strip the [5] explicit
	// wrapper when unmarshaling into asn1.RawValue — the outer tag stays in the RawValue
	// itself, and Bytes holds the content of that outer tag = the full APPLICATION[1]).
	if _, err := r.Ticket.Unmarshal(inner.Ticket.Bytes); err != nil {
		return 0, fmt.Errorf("tgsrep ticket unmarshal: %w", err)
	}

	return consumed, nil
}
