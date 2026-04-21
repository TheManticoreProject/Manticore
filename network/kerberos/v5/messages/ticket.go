package messages

import (
	"encoding/asn1"
)

// ticketInner is the inner SEQUENCE of a Kerberos Ticket for unmarshaling.
type ticketInner struct {
	// TktVno is the ticket version number (always 5).
	TktVno int `asn1:"explicit,tag:0"`
	// Realm is the realm of the principal named in the SName field.
	Realm string `asn1:"explicit,tag:1,generalstring"`
	// SName is the name of the server for which the ticket was issued.
	SName PrincipalName `asn1:"explicit,tag:2"`
	// EncPart is the encrypted part of the ticket.
	EncPart EncryptedData `asn1:"explicit,tag:3"`
}

// ticketMarshal is the wire representation of Ticket for marshaling.
// Realm has no struct tag: Go ignores explicit,tag:N for asn1.RawValue,
// so realmExplicit() pre-builds the [1] EXPLICIT { GeneralString } wrapper.
type ticketMarshal struct {
	TktVno  int                  `asn1:"explicit,tag:0"`
	Realm   asn1.RawValue        // pre-encoded [1] EXPLICIT { GeneralString } by realmExplicit
	SName   PrincipalNameMarshal `asn1:"explicit,tag:2"`
	EncPart EncryptedData        `asn1:"explicit,tag:3"`
}

// Ticket is a Kerberos ticket (APPLICATION[1]), as defined in RFC 4120 Section 5.3.
// It carries an encrypted session key and authorization data for a service principal.
type Ticket struct {
	// TktVno is the Kerberos version number embedded in the ticket (always 5).
	TktVno int
	// Realm is the realm of the service principal.
	Realm string
	// SName is the name of the service principal.
	SName PrincipalName
	// EncPart is the encrypted portion of the ticket.
	EncPart EncryptedData
}

// Marshal encodes the Ticket as an ASN.1 APPLICATION[1] wrapped SEQUENCE.
func (t *Ticket) Marshal() ([]byte, error) {
	inner := ticketMarshal{
		TktVno:  t.TktVno,
		Realm:   realmExplicit(1, t.Realm),
		SName:   MarshalPrincipalName(t.SName),
		EncPart: t.EncPart,
	}
	seq_bytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(1, seq_bytes)
}

// Unmarshal decodes a Ticket from an ASN.1 APPLICATION[1] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (t *Ticket) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, 1)
	if err != nil {
		return 0, err
	}

	var inner ticketInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, err
	}

	t.TktVno = inner.TktVno
	t.Realm = inner.Realm
	t.SName = inner.SName
	t.EncPart = inner.EncPart
	return consumed, nil
}
