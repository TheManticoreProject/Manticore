package messages

import (
	"encoding/asn1"
	"fmt"
)

// kdcRepInner is the inner SEQUENCE of a KDC reply (AS-REP or TGS-REP).
type kdcRepInner struct {
	// PVNO is the Kerberos protocol version (always 5).
	PVNO int `asn1:"explicit,tag:0"`
	// MsgType is the message type (MsgTypeASRep = 11 or MsgTypeTGSRep = 13).
	MsgType int `asn1:"explicit,tag:1"`
	// PAData contains optional pre-authentication data.
	PAData []PAData `asn1:"explicit,tag:2,optional"`
	// CRealm is the client's realm.
	CRealm string `asn1:"explicit,tag:3,generalstring"`
	// CName is the client's principal name.
	CName PrincipalName `asn1:"explicit,tag:4"`
	// Ticket is pre-encoded as [5] EXPLICIT { APPLICATION[1] bytes }.
	// Go ignores explicit,tag:N for asn1.RawValue (both Marshal and Unmarshal), so we store
	// the [5] context wrapper ourselves. Bytes = APPLICATION[1] TLV after Unmarshal.
	Ticket asn1.RawValue
	// EncPart is the encrypted part of the reply containing the session key.
	EncPart EncryptedData `asn1:"explicit,tag:6"`
}

// ASRep is a Kerberos AS-REP (Authentication Service Reply) message,
// APPLICATION[11], as defined in RFC 4120 Section 5.4.2.
// It is sent by the KDC in response to a successful AS-REQ.
type ASRep struct {
	// PVNO is the Kerberos protocol version (always 5).
	PVNO int
	// MsgType is the message type (always MsgTypeASRep = 11).
	MsgType int
	// PAData contains pre-authentication data (rarely set in AS-REP).
	PAData []PAData
	// CRealm is the realm of the client.
	CRealm string
	// CName is the client's principal name as returned by the KDC.
	CName PrincipalName
	// Ticket is the issued Ticket Granting Ticket (parsed).
	Ticket Ticket
	// TicketRaw holds the raw APPLICATION[1] ticket bytes as received from the KDC.
	// Use these verbatim in AP-REQ to avoid re-encoding differences.
	TicketRaw []byte
	// EncPart is the encrypted reply body, decryptable with the client's key.
	EncPart EncryptedData
}

// Marshal encodes the AS-REP as an ASN.1 APPLICATION[11] wrapped SEQUENCE.
func (r *ASRep) Marshal() ([]byte, error) {
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
		MsgType: MsgTypeASRep,
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
	return wrapApplication(MsgTypeASRep, seq_bytes)
}

// Unmarshal decodes an AS-REP from an ASN.1 APPLICATION[11] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (r *ASRep) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, MsgTypeASRep)
	if err != nil {
		return 0, fmt.Errorf("asrep: %w", err)
	}

	var inner kdcRepInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("asrep inner unmarshal: %w", err)
	}

	r.PVNO = inner.PVNO
	r.MsgType = inner.MsgType
	r.PAData = inner.PAData
	r.CRealm = inner.CRealm
	r.CName = inner.CName
	r.EncPart = inner.EncPart

	// inner.Ticket.Bytes holds the APPLICATION[1] ticket bytes as sent by the KDC.
	// (Go does not strip the [5] explicit wrapper for asn1.RawValue fields — the outer
	// context tag stays in the RawValue, and Bytes = content inside that tag = APPLICATION[1].)
	r.TicketRaw = inner.Ticket.Bytes
	if _, err := r.Ticket.Unmarshal(r.TicketRaw); err != nil {
		return 0, fmt.Errorf("asrep ticket unmarshal: %w", err)
	}

	return consumed, nil
}
