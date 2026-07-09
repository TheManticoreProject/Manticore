package messages

import (
	"encoding/asn1"
	"time"
)

// TransitedEncoding is the transited-realm field of a ticket, as defined in
// RFC 4120 Section 5.3. A freshly issued (or forged) ticket carries an empty
// contents with TRType 0 (DOMAIN-X500-COMPRESS), meaning no cross-realm hops.
type TransitedEncoding struct {
	// TRType identifies the encoding of the transited field (0 = X.500 compress).
	TRType int `asn1:"explicit,tag:0"`
	// Contents holds the encoded transited realms (empty when no realms transited).
	Contents []byte `asn1:"explicit,tag:1"`
}

// EncTicketPart is the encrypted portion of a Kerberos ticket (APPLICATION[3]),
// as defined in RFC 4120 Section 5.3. It carries the ticket flags, the session
// key, the client principal, validity times, and — for a Windows ticket — the
// PAC inside the authorization-data field. The KDC encrypts its DER encoding
// under the service's long-term key (key usage 2); forging a ticket means
// building this structure and encrypting it under a compromised service or
// krbtgt key.
type EncTicketPart struct {
	// Flags are the ticket flags (forwardable, renewable, pre-authent, …).
	Flags asn1.BitString
	// Key is the session key sealed inside the ticket.
	Key EncryptionKey
	// CRealm is the client's realm.
	CRealm string
	// CName is the client principal the ticket is issued to.
	CName PrincipalName
	// Transited is the transited-realm encoding (empty for a locally issued ticket).
	Transited TransitedEncoding
	// AuthTime is the time of the initial authentication.
	AuthTime time.Time
	// StartTime is the time from which the ticket is valid (optional).
	StartTime time.Time
	// EndTime is the ticket's expiry time.
	EndTime time.Time
	// RenewTill is the end of the renewable lifetime (optional).
	RenewTill time.Time
	// AuthorizationData carries the authorization-data elements (the AD-IF-RELEVANT
	// wrapped AD-WIN2K-PAC for a Windows ticket). Optional.
	AuthorizationData []AuthorizationData
}

// encTicketPartMarshal is the wire representation used for marshaling. As with
// ticketMarshal, CRealm is pre-encoded as a [2] EXPLICIT { GeneralString }
// RawValue (Go's asn1 would otherwise emit PrintableString), and CName uses the
// GeneralString-based PrincipalNameMarshal.
type encTicketPartMarshal struct {
	Flags             asn1.BitString       `asn1:"explicit,tag:0"`
	Key               EncryptionKey        `asn1:"explicit,tag:1"`
	CRealm            asn1.RawValue        // pre-encoded [2] EXPLICIT { GeneralString }
	CName             PrincipalNameMarshal `asn1:"explicit,tag:3"`
	Transited         TransitedEncoding    `asn1:"explicit,tag:4"`
	AuthTime          time.Time            `asn1:"explicit,tag:5,generalized"`
	StartTime         time.Time            `asn1:"explicit,tag:6,optional,generalized"`
	EndTime           time.Time            `asn1:"explicit,tag:7,generalized"`
	RenewTill         time.Time            `asn1:"explicit,tag:8,optional,generalized"`
	AuthorizationData []AuthorizationData  `asn1:"explicit,tag:10,optional"`
}

// Marshal encodes the EncTicketPart as an ASN.1 APPLICATION[3] wrapped SEQUENCE,
// ready to be encrypted under the service (or krbtgt) key as a ticket enc-part.
func (e *EncTicketPart) Marshal() ([]byte, error) {
	inner := encTicketPartMarshal{
		Flags:     e.Flags,
		Key:       e.Key,
		CRealm:    realmExplicit(2, e.CRealm),
		CName:     MarshalPrincipalName(e.CName),
		Transited: e.Transited,
		AuthTime:  normalizeTime(e.AuthTime),
		EndTime:   normalizeTime(e.EndTime),
	}
	if !e.StartTime.IsZero() {
		inner.StartTime = normalizeTime(e.StartTime)
	}
	if !e.RenewTill.IsZero() {
		inner.RenewTill = normalizeTime(e.RenewTill)
	}
	inner.AuthorizationData = e.AuthorizationData

	seqBytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(3, seqBytes)
}

// Unmarshal decodes an EncTicketPart from an ASN.1 APPLICATION[3] wrapped
// SEQUENCE. Returns the number of bytes consumed from data.
func (e *EncTicketPart) Unmarshal(data []byte) (int, error) {
	innerBytes, consumed, err := unwrapApplication(data, 3)
	if err != nil {
		return 0, err
	}

	var inner struct {
		Flags             asn1.BitString      `asn1:"explicit,tag:0"`
		Key               EncryptionKey       `asn1:"explicit,tag:1"`
		CRealm            string              `asn1:"explicit,tag:2,generalstring"`
		CName             PrincipalName       `asn1:"explicit,tag:3"`
		Transited         TransitedEncoding   `asn1:"explicit,tag:4"`
		AuthTime          time.Time           `asn1:"explicit,tag:5,generalized"`
		StartTime         time.Time           `asn1:"explicit,tag:6,optional,generalized"`
		EndTime           time.Time           `asn1:"explicit,tag:7,generalized"`
		RenewTill         time.Time           `asn1:"explicit,tag:8,optional,generalized"`
		CAddr             asn1.RawValue       `asn1:"explicit,tag:9,optional"`
		AuthorizationData []AuthorizationData `asn1:"explicit,tag:10,optional"`
	}
	if _, err := asn1.Unmarshal(innerBytes, &inner); err != nil {
		return 0, err
	}

	e.Flags = inner.Flags
	e.Key = inner.Key
	e.CRealm = inner.CRealm
	e.CName = inner.CName
	e.Transited = inner.Transited
	e.AuthTime = inner.AuthTime
	e.StartTime = inner.StartTime
	e.EndTime = inner.EndTime
	e.RenewTill = inner.RenewTill
	e.AuthorizationData = inner.AuthorizationData
	return consumed, nil
}
