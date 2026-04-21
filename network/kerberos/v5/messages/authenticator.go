package messages

import (
	"encoding/asn1"
	"fmt"
	"time"
)

// Checksum contains a cryptographic checksum as defined in RFC 4120 Section 5.2.9.
type Checksum struct {
	// CKSumType identifies the checksum algorithm.
	CKSumType int `asn1:"explicit,tag:0"`
	// Checksum contains the raw checksum bytes.
	Checksum []byte `asn1:"explicit,tag:1"`
}

// EncryptionKey holds a Kerberos encryption key as defined in RFC 4120 Section 5.2.9.
type EncryptionKey struct {
	// KeyType identifies the encryption algorithm.
	KeyType int `asn1:"explicit,tag:0"`
	// KeyValue contains the raw key bytes.
	KeyValue []byte `asn1:"explicit,tag:1"`
}

// authenticatorInner is the inner SEQUENCE of a Kerberos Authenticator for unmarshaling.
type authenticatorInner struct {
	// AVno is the Authenticator version number (always 5).
	AVno int `asn1:"explicit,tag:0"`
	// CRealm is the client's realm.
	CRealm string `asn1:"explicit,tag:1,generalstring"`
	// CName is the client's principal name.
	CName PrincipalName `asn1:"explicit,tag:2"`
	// Cksum is an optional checksum of the application data.
	Cksum Checksum `asn1:"explicit,tag:3,optional"`
	// CUSec is the microseconds component of the client timestamp.
	CUSec int `asn1:"explicit,tag:4"`
	// CTime is the client timestamp (used to detect replays).
	CTime time.Time `asn1:"explicit,tag:5,generalized"`
	// SubKey is an optional sub-session key chosen by the client.
	SubKey EncryptionKey `asn1:"explicit,tag:6,optional"`
	// SeqNumber is an optional sequence number for ordering messages.
	SeqNumber int `asn1:"explicit,tag:7,optional"`
	// AuthorizationData contains optional authorization data.
	AuthorizationData []AuthorizationData `asn1:"explicit,tag:8,optional"`
}

// authenticatorMarshal is the wire representation for marshaling.
// CRealm uses realmExplicit; CName uses PrincipalNameMarshal for GeneralString encoding.
type authenticatorMarshal struct {
	AVno              int                  `asn1:"explicit,tag:0"`
	CRealm            asn1.RawValue        // pre-encoded [1] EXPLICIT { GeneralString }
	CName             PrincipalNameMarshal `asn1:"explicit,tag:2"`
	Cksum             Checksum             `asn1:"explicit,tag:3,optional"`
	CUSec             int                  `asn1:"explicit,tag:4"`
	CTime             time.Time            `asn1:"explicit,tag:5,generalized"`
	SubKey            EncryptionKey        `asn1:"explicit,tag:6,optional"`
	SeqNumber         int                  `asn1:"explicit,tag:7,optional"`
	AuthorizationData []AuthorizationData  `asn1:"explicit,tag:8,optional"`
}

// Authenticator is a Kerberos Authenticator (APPLICATION[2]),
// as defined in RFC 4120 Section 5.5.1.
// It is encrypted within an AP-REQ and proves the client's identity.
type Authenticator struct {
	// AVno is the Authenticator version number (always 5).
	AVno int
	// CRealm is the realm of the client.
	CRealm string
	// CName is the client's principal name.
	CName PrincipalName
	// CUSec is the microseconds component of CTime.
	CUSec int
	// CTime is the client's current time (must match server time within clock skew).
	CTime time.Time
	// SubKey is an optional client-chosen sub-session key.
	SubKey *EncryptionKey
	// SeqNumber is the optional sequence number.
	SeqNumber int
}

// Marshal encodes the Authenticator as an ASN.1 APPLICATION[2] wrapped SEQUENCE.
func (a *Authenticator) Marshal() ([]byte, error) {
	inner := authenticatorMarshal{
		AVno:      KerberosV5,
		CRealm:    realmExplicit(1, a.CRealm),
		CName:     MarshalPrincipalName(a.CName),
		CUSec:     a.CUSec,
		CTime:     a.CTime,
		SeqNumber: a.SeqNumber,
	}
	if a.SubKey != nil {
		inner.SubKey = *a.SubKey
	}

	seq_bytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(2, seq_bytes)
}

// Unmarshal decodes an Authenticator from an ASN.1 APPLICATION[2] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (a *Authenticator) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, 2)
	if err != nil {
		return 0, fmt.Errorf("authenticator: %w", err)
	}

	var inner authenticatorInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("authenticator inner unmarshal: %w", err)
	}

	a.AVno = inner.AVno
	a.CRealm = inner.CRealm
	a.CName = inner.CName
	a.CUSec = inner.CUSec
	a.CTime = inner.CTime
	a.SeqNumber = inner.SeqNumber
	// SubKey is optional; only set if KeyType is non-zero
	if inner.SubKey.KeyType != 0 {
		sk := inner.SubKey
		a.SubKey = &sk
	}

	return consumed, nil
}
