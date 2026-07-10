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

// encRepPartMarshal is the wire form of an EncKDCRepPart for marshaling. Unlike
// encRepPartInner it encodes SRealm as a GeneralString and SName via
// PrincipalNameMarshal, which Go's encoding/asn1 would otherwise emit as
// PrintableString (RFC 4120 requires GeneralString). KeyExpiration [3] is not
// exposed by the public structs and is always omitted.
type encRepPartMarshal struct {
	Key       EncryptionKey        `asn1:"explicit,tag:0"`
	LastReq   []LastReq            `asn1:"explicit,tag:1"`
	Nonce     int                  `asn1:"explicit,tag:2"`
	Flags     asn1.BitString       `asn1:"explicit,tag:4"`
	AuthTime  time.Time            `asn1:"explicit,tag:5,generalized"`
	StartTime time.Time            `asn1:"explicit,tag:6,optional,generalized"`
	EndTime   time.Time            `asn1:"explicit,tag:7,generalized"`
	RenewTill time.Time            `asn1:"explicit,tag:8,optional,generalized"`
	SRealm    asn1.RawValue        // pre-encoded [9] EXPLICIT { GeneralString }
	SName     PrincipalNameMarshal `asn1:"explicit,tag:10"`
}

// marshalEncRepPart assembles an EncKDCRepPart and wraps it in the given
// APPLICATION tag (25 for AS-REP, 26 for TGS-REP). Optional times are emitted
// only when non-zero.
func marshalEncRepPart(appTag int, key EncryptionKey, nonce int, flags asn1.BitString,
	authTime, startTime, endTime, renewTill time.Time, srealm string, sname PrincipalName) ([]byte, error) {
	inner := encRepPartMarshal{
		Key:      key,
		LastReq:  []LastReq{},
		Nonce:    nonce,
		Flags:    flags,
		AuthTime: normalizeTime(authTime),
		EndTime:  normalizeTime(endTime),
		SRealm:   realmExplicit(9, srealm),
		SName:    MarshalPrincipalName(sname),
	}
	if !startTime.IsZero() {
		inner.StartTime = normalizeTime(startTime)
	}
	if !renewTill.IsZero() {
		inner.RenewTill = normalizeTime(renewTill)
	}
	seq_bytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(appTag, seq_bytes)
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

// Marshal encodes the EncASRepPart as an ASN.1 APPLICATION[25] wrapped SEQUENCE.
func (e *EncASRepPart) Marshal() ([]byte, error) {
	return marshalEncRepPart(25, e.Key, e.Nonce, e.Flags, e.AuthTime, e.StartTime, e.EndTime, e.RenewTill, e.SRealm, e.SName)
}

// Unmarshal decodes an EncASRepPart from an ASN.1 APPLICATION[25] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (e *EncASRepPart) Unmarshal(data []byte) (int, error) {
	// RFC 4120 specifies APPLICATION[25] here, but some KDCs tag an AS-REP
	// enc-part as APPLICATION[26] (EncTGSRepPart); accept both for interop.
	inner_bytes, consumed, err := unwrapApplicationOneOf(data, 25, 26)
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

// Marshal encodes the EncTGSRepPart as an ASN.1 APPLICATION[26] wrapped SEQUENCE.
func (e *EncTGSRepPart) Marshal() ([]byte, error) {
	return marshalEncRepPart(26, e.Key, e.Nonce, e.Flags, e.AuthTime, e.StartTime, e.EndTime, e.RenewTill, e.SRealm, e.SName)
}

// Unmarshal decodes an EncTGSRepPart from an ASN.1 APPLICATION[26] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (e *EncTGSRepPart) Unmarshal(data []byte) (int, error) {
	// RFC 4120 specifies APPLICATION[26]; accept APPLICATION[25] as well for
	// symmetry with lenient KDC enc-part tagging.
	inner_bytes, consumed, err := unwrapApplicationOneOf(data, 26, 25)
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
