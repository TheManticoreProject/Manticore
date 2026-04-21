package messages

import (
	"encoding/asn1"
	"fmt"
	"time"
)

// LastReq is a last-request entry as defined in RFC 4120 Section 5.4.2.
type LastReq struct {
	// LRType identifies the type of last request.
	LRType int `asn1:"explicit,tag:0"`
	// LRValue is the time of the last request.
	LRValue time.Time `asn1:"explicit,tag:1,generalized"`
}

// encRepPartInner is the inner SEQUENCE shared by EncASRepPart and EncTGSRepPart.
type encRepPartInner struct {
	Key           EncryptionKey  `asn1:"explicit,tag:0"`
	LastReq       []LastReq      `asn1:"explicit,tag:1"`
	Nonce         int            `asn1:"explicit,tag:2"`
	KeyExpiration time.Time      `asn1:"explicit,tag:3,optional,generalized"`
	Flags         asn1.BitString `asn1:"explicit,tag:4"`
	AuthTime      time.Time      `asn1:"explicit,tag:5,generalized"`
	StartTime     time.Time      `asn1:"explicit,tag:6,optional,generalized"`
	EndTime       time.Time      `asn1:"explicit,tag:7,generalized"`
	RenewTill     time.Time      `asn1:"explicit,tag:8,optional,generalized"`
	SRealm        string         `asn1:"explicit,tag:9,generalstring"`
	SName         PrincipalName  `asn1:"explicit,tag:10"`
}

// EncASRepPart is the decrypted enc-part of an AS-REP (APPLICATION 25),
// as defined in RFC 4120 Section 5.4.2.
// It contains the session key and ticket metadata.
type EncASRepPart struct {
	// Key is the session key for use with the issued ticket.
	Key EncryptionKey
	// Nonce must match the nonce in the AS-REQ.
	Nonce int
	// Flags contains the ticket flags.
	Flags asn1.BitString
	// AuthTime is the time the client was authenticated.
	AuthTime time.Time
	// StartTime is the ticket's start time (optional).
	StartTime time.Time
	// EndTime is the ticket's expiry time.
	EndTime time.Time
	// RenewTill is the renewable lifetime end time (optional).
	RenewTill time.Time
	// SRealm is the realm of the service.
	SRealm string
	// SName is the service principal name.
	SName PrincipalName
}

// Unmarshal decodes an EncASRepPart from an ASN.1 APPLICATION[25] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (e *EncASRepPart) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, 25)
	if err != nil {
		return 0, fmt.Errorf("encasreppart: %w", err)
	}

	var inner encRepPartInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("encasreppart inner unmarshal: %w", err)
	}

	e.Key = inner.Key
	e.Nonce = inner.Nonce
	e.Flags = inner.Flags
	e.AuthTime = inner.AuthTime
	e.StartTime = inner.StartTime
	e.EndTime = inner.EndTime
	e.RenewTill = inner.RenewTill
	e.SRealm = inner.SRealm
	e.SName = inner.SName
	return consumed, nil
}

// EncTGSRepPart is the decrypted enc-part of a TGS-REP (APPLICATION 26),
// as defined in RFC 4120 Section 5.4.2.
// It has the same structure as EncASRepPart but a different APPLICATION tag.
type EncTGSRepPart struct {
	// Key is the session key for use with the service ticket.
	Key EncryptionKey
	// Nonce must match the nonce in the TGS-REQ.
	Nonce int
	// Flags contains the ticket flags.
	Flags asn1.BitString
	// AuthTime is the time of original authentication.
	AuthTime time.Time
	// StartTime is the ticket's start time (optional).
	StartTime time.Time
	// EndTime is the ticket's expiry time.
	EndTime time.Time
	// RenewTill is the renewable lifetime end time (optional).
	RenewTill time.Time
	// SRealm is the realm of the service.
	SRealm string
	// SName is the service principal name.
	SName PrincipalName
}

// Unmarshal decodes an EncTGSRepPart from an ASN.1 APPLICATION[26] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (e *EncTGSRepPart) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, 26)
	if err != nil {
		return 0, fmt.Errorf("enctgsreppart: %w", err)
	}

	var inner encRepPartInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("enctgsreppart inner unmarshal: %w", err)
	}

	e.Key = inner.Key
	e.Nonce = inner.Nonce
	e.Flags = inner.Flags
	e.AuthTime = inner.AuthTime
	e.StartTime = inner.StartTime
	e.EndTime = inner.EndTime
	e.RenewTill = inner.RenewTill
	e.SRealm = inner.SRealm
	e.SName = inner.SName
	return consumed, nil
}
