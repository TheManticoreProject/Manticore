package messages

import (
	"encoding/asn1"
	"time"
)

// kdcReqBodyMarshal is the wire representation of KDCReqBody for marshaling.
// It uses GeneralString (asn1.RawValue) for Realm and PrincipalNameMarshal for names,
// because Go's asn1 ignores the "generalstring" tag and produces PrintableString.
//
// Realm has no struct tag: Go ignores explicit,tag:N for asn1.RawValue fields, so
// realmExplicit() pre-builds the [2] EXPLICIT { GeneralString } wrapper instead.
type kdcReqBodyMarshal struct {
	KDCOptions  asn1.BitString       `asn1:"explicit,tag:0"`
	CName       PrincipalNameMarshal `asn1:"explicit,tag:1,optional"`
	Realm       asn1.RawValue        // pre-encoded as [2] EXPLICIT { GeneralString } by realmExplicit
	SName       PrincipalNameMarshal `asn1:"explicit,tag:3,optional"`
	From        time.Time            `asn1:"explicit,tag:4,optional,generalized"`
	Till        time.Time            `asn1:"explicit,tag:5,generalized"`
	RTime       time.Time            `asn1:"explicit,tag:6,optional,generalized"`
	Nonce       int                  `asn1:"explicit,tag:7"`
	EType       []int                `asn1:"explicit,tag:8"`
	Addresses   []HostAddress        `asn1:"explicit,tag:9,optional"`
	EncAuthData EncryptedData        `asn1:"explicit,tag:10,optional"`
	// additional-tickets [11] is NOT a struct field: Go's encoding/asn1 would
	// emit each Ticket as a bare SEQUENCE instead of its APPLICATION[1] form.
	// It is spliced in by encodeKDCReqBodyForTGS when present.
}

// marshalKDCReqBody converts a KDCReqBody to its GeneralString-encoded form
// (fields 0-10; additional-tickets are handled separately).
func marshalKDCReqBody(b KDCReqBody) kdcReqBodyMarshal {
	return kdcReqBodyMarshal{
		KDCOptions:  b.KDCOptions,
		CName:       MarshalPrincipalName(b.CName),
		Realm:       realmExplicit(2, b.Realm),
		SName:       MarshalPrincipalName(b.SName),
		From:        b.From,
		Till:        b.Till,
		RTime:       b.RTime,
		Nonce:       b.Nonce,
		EType:       b.EType,
		Addresses:   b.Addresses,
		EncAuthData: b.EncAuthData,
	}
}

// encodeKDCReqBodyForTGS marshals a KDC-REQ-BODY and, when the body carries
// additional-tickets, splices in a correctly-encoded [11] EXPLICIT SEQUENCE OF
// Ticket (each an APPLICATION[1] TLV). AdditTicketsRaw is preferred (verbatim
// KDC-issued bytes); otherwise AdditTickets are marshaled. Returns the complete
// KDC-REQ-BODY SEQUENCE TLV.
func encodeKDCReqBodyForTGS(b KDCReqBody) ([]byte, error) {
	base, err := asn1.Marshal(marshalKDCReqBody(b))
	if err != nil {
		return nil, err
	}
	if len(b.AdditTicketsRaw) == 0 && len(b.AdditTickets) == 0 {
		return base, nil
	}

	var tktTLVs []byte
	for _, raw := range b.AdditTicketsRaw {
		tktTLVs = append(tktTLVs, raw...)
	}
	for i := range b.AdditTickets {
		m, err := b.AdditTickets[i].Marshal()
		if err != nil {
			return nil, err
		}
		tktTLVs = append(tktTLVs, m...)
	}
	seqOf, err := asn1.Marshal(asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSequence, IsCompound: true, Bytes: tktTLVs})
	if err != nil {
		return nil, err
	}
	elem11, err := asn1.Marshal(asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 11, IsCompound: true, Bytes: seqOf})
	if err != nil {
		return nil, err
	}

	// Splice elem11 into the base body SEQUENCE's content.
	var body asn1.RawValue
	if _, err := asn1.Unmarshal(base, &body); err != nil {
		return nil, err
	}
	content := make([]byte, 0, len(body.Bytes)+len(elem11))
	content = append(content, body.Bytes...)
	content = append(content, elem11...)
	return asn1.Marshal(asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSequence, IsCompound: true, Bytes: content})
}

// KDCReqBody is the body of a KDC request (AS-REQ or TGS-REQ),
// as defined in RFC 4120 Section 5.4.1.
type KDCReqBody struct {
	// KDCOptions contains bit flags controlling the KDC request behavior.
	KDCOptions asn1.BitString `asn1:"explicit,tag:0"`
	// CName is the client principal name (present in AS-REQ, absent in TGS-REQ).
	CName PrincipalName `asn1:"explicit,tag:1,optional"`
	// Realm is the realm for the request (crealm in AS-REQ, srealm in TGS-REQ).
	Realm string `asn1:"explicit,tag:2,generalstring"`
	// SName is the server principal name being requested.
	SName PrincipalName `asn1:"explicit,tag:3,optional"`
	// From is the requested start time for the ticket (optional).
	From time.Time `asn1:"explicit,tag:4,optional,generalized"`
	// Till is the requested expiry time for the ticket.
	Till time.Time `asn1:"explicit,tag:5,generalized"`
	// RTime is the requested renewable lifetime end time (optional).
	RTime time.Time `asn1:"explicit,tag:6,optional,generalized"`
	// Nonce is a random number used to detect replays.
	Nonce int `asn1:"explicit,tag:7"`
	// EType lists the client's supported encryption types, in preference order.
	EType []int `asn1:"explicit,tag:8"`
	// Addresses restricts the ticket to specific network addresses (optional).
	Addresses []HostAddress `asn1:"explicit,tag:9,optional"`
	// EncAuthData contains encrypted authorization data (optional, TGS-REQ).
	EncAuthData EncryptedData `asn1:"explicit,tag:10,optional"`
	// AdditTickets holds additional tickets (parsed form) for the TGS-REQ
	// additional-tickets field — used by U2U and S4U2Proxy. Marshaled by
	// encodeKDCReqBodyForTGS (not by generic asn1, which would mis-encode the
	// APPLICATION[1] tickets), hence asn1:"-".
	AdditTickets []Ticket `asn1:"-"`
	// AdditTicketsRaw holds the raw APPLICATION[1] bytes of additional tickets,
	// preferred over AdditTickets on marshal to re-emit KDC-issued bytes verbatim.
	AdditTicketsRaw [][]byte `asn1:"-"`
}
