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
	KDCOptions   asn1.BitString       `asn1:"explicit,tag:0"`
	CName        PrincipalNameMarshal `asn1:"explicit,tag:1,optional"`
	Realm        asn1.RawValue        // pre-encoded as [2] EXPLICIT { GeneralString } by realmExplicit
	SName        PrincipalNameMarshal `asn1:"explicit,tag:3,optional"`
	From         time.Time            `asn1:"explicit,tag:4,optional,generalized"`
	Till         time.Time            `asn1:"explicit,tag:5,generalized"`
	RTime        time.Time            `asn1:"explicit,tag:6,optional,generalized"`
	Nonce        int                  `asn1:"explicit,tag:7"`
	EType        []int                `asn1:"explicit,tag:8"`
	Addresses    []HostAddress        `asn1:"explicit,tag:9,optional"`
	EncAuthData  EncryptedData        `asn1:"explicit,tag:10,optional"`
	AdditTickets []Ticket             `asn1:"explicit,tag:11,optional"`
}

// marshalKDCReqBody converts a KDCReqBody to its GeneralString-encoded form.
func marshalKDCReqBody(b KDCReqBody) kdcReqBodyMarshal {
	return kdcReqBodyMarshal{
		KDCOptions:   b.KDCOptions,
		CName:        MarshalPrincipalName(b.CName),
		Realm:        realmExplicit(2, b.Realm),
		SName:        MarshalPrincipalName(b.SName),
		From:         b.From,
		Till:         b.Till,
		RTime:        b.RTime,
		Nonce:        b.Nonce,
		EType:        b.EType,
		Addresses:    b.Addresses,
		EncAuthData:  b.EncAuthData,
		AdditTickets: b.AdditTickets,
	}
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
	// AdditTickets contains additional tickets (optional, for TGS renewal/forwarding).
	AdditTickets []Ticket `asn1:"explicit,tag:11,optional"`
}
