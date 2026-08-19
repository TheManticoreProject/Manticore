package messages

import (
	"encoding/asn1"
	"fmt"
	"time"
)

// KRB-CRED and its encrypted part carry delegated/exported tickets between
// principals (RFC 4120 Sections 5.8.1). They are the on-the-wire form behind
// the ".kirbi" credential files used for pass-the-ticket, so both marshal and
// unmarshal are implemented.
//
// Marshaling is done with a small DER builder (derSequence/derExplicit +
// asn1.MarshalWithParams) rather than plain struct marshaling because the
// realm fields are GeneralString and several fields are OPTIONAL — Go's
// encoding/asn1 emits PrintableString for Go strings and cannot omit a zero
// asn1.RawValue, so field-by-field assembly is required. Unmarshaling can use
// ordinary struct tags (Go decodes GeneralString correctly on the way in).

// derSequence wraps the concatenation of the given DER elements in a universal
// SEQUENCE TLV.
func derSequence(elements ...[]byte) ([]byte, error) {
	var body []byte
	for _, e := range elements {
		body = append(body, e...)
	}
	return asn1.Marshal(asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSequence, IsCompound: true, Bytes: body})
}

// derExplicit wraps an already-encoded inner TLV in an [tag] EXPLICIT
// context-specific constructed element.
func derExplicit(tag int, innerTLV []byte) ([]byte, error) {
	return asn1.Marshal(asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: tag, IsCompound: true, Bytes: innerTLV})
}

// KrbCredInfo is one entry of an EncKrbCredPart's ticket-info (RFC 4120
// Section 5.8.1). Every field except Key is OPTIONAL on the wire; in practice
// a ticket carries prealm/pname/flags/times/srealm/sname.
type KrbCredInfo struct {
	Key       EncryptionKey
	PRealm    string
	PName     PrincipalName
	Flags     asn1.BitString
	AuthTime  time.Time
	StartTime time.Time
	EndTime   time.Time
	RenewTill time.Time
	SRealm    string
	SName     PrincipalName
}

// krbCredInfoInner is the unmarshal view of a KrbCredInfo SEQUENCE.
type krbCredInfoInner struct {
	Key       EncryptionKey  `asn1:"explicit,tag:0"`
	PRealm    string         `asn1:"explicit,tag:1,optional,generalstring"`
	PName     PrincipalName  `asn1:"explicit,tag:2,optional"`
	Flags     asn1.BitString `asn1:"explicit,tag:3,optional"`
	AuthTime  time.Time      `asn1:"explicit,tag:4,optional,generalized"`
	StartTime time.Time      `asn1:"explicit,tag:5,optional,generalized"`
	EndTime   time.Time      `asn1:"explicit,tag:6,optional,generalized"`
	RenewTill time.Time      `asn1:"explicit,tag:7,optional,generalized"`
	SRealm    string         `asn1:"explicit,tag:8,optional,generalstring"`
	SName     PrincipalName  `asn1:"explicit,tag:9,optional"`
	CAddr     []HostAddress  `asn1:"explicit,tag:10,optional"`
}

// marshal encodes a single KrbCredInfo SEQUENCE, emitting only the present
// (non-zero) optional fields.
func (info *KrbCredInfo) marshal() ([]byte, error) {
	var elems [][]byte

	key, err := asn1.MarshalWithParams(info.Key, "explicit,tag:0")
	if err != nil {
		return nil, err
	}
	elems = append(elems, key)

	if info.PRealm != "" {
		elems = append(elems, mustMarshal(realmExplicit(1, info.PRealm)))
	}
	if len(info.PName.NameString) > 0 {
		p, err := asn1.MarshalWithParams(MarshalPrincipalName(info.PName), "explicit,tag:2")
		if err != nil {
			return nil, err
		}
		elems = append(elems, p)
	}
	if info.Flags.BitLength > 0 {
		f, err := asn1.MarshalWithParams(info.Flags, "explicit,tag:3")
		if err != nil {
			return nil, err
		}
		elems = append(elems, f)
	}
	// Times are emitted in tag order (DER requires ascending context tags).
	if tv, err := marshalOptTime(4, info.AuthTime); err != nil {
		return nil, err
	} else if tv != nil {
		elems = append(elems, tv)
	}
	if tv, err := marshalOptTime(5, info.StartTime); err != nil {
		return nil, err
	} else if tv != nil {
		elems = append(elems, tv)
	}
	if tv, err := marshalOptTime(6, info.EndTime); err != nil {
		return nil, err
	} else if tv != nil {
		elems = append(elems, tv)
	}
	if tv, err := marshalOptTime(7, info.RenewTill); err != nil {
		return nil, err
	} else if tv != nil {
		elems = append(elems, tv)
	}
	if info.SRealm != "" {
		elems = append(elems, mustMarshal(realmExplicit(8, info.SRealm)))
	}
	if len(info.SName.NameString) > 0 {
		s, err := asn1.MarshalWithParams(MarshalPrincipalName(info.SName), "explicit,tag:9")
		if err != nil {
			return nil, err
		}
		elems = append(elems, s)
	}
	return derSequence(elems...)
}

// marshalOptTime returns the [tag] EXPLICIT GeneralizedTime encoding of t, or
// (nil, nil) if t is the zero time.
func marshalOptTime(tag int, t time.Time) ([]byte, error) {
	if t.IsZero() {
		return nil, nil
	}
	return asn1.MarshalWithParams(normalizeTime(t), fmt.Sprintf("explicit,tag:%d,generalized", tag))
}

// mustMarshal marshals a RawValue that is known to be well-formed (the realm
// GeneralString wrappers), panicking only on a programming error.
func mustMarshal(rv asn1.RawValue) []byte {
	b, err := asn1.Marshal(rv)
	if err != nil {
		panic("messages: marshal RawValue: " + err.Error())
	}
	return b
}

func (info *krbCredInfoInner) toPublic() KrbCredInfo {
	return KrbCredInfo{
		Key:       info.Key,
		PRealm:    info.PRealm,
		PName:     info.PName,
		Flags:     info.Flags,
		AuthTime:  info.AuthTime,
		StartTime: info.StartTime,
		EndTime:   info.EndTime,
		RenewTill: info.RenewTill,
		SRealm:    info.SRealm,
		SName:     info.SName,
	}
}

// EncKrbCredPart is the decrypted enc-part of a KRB-CRED (APPLICATION[29]),
// RFC 4120 Section 5.8.1. It is encrypted under a key the two parties share
// (key usage 14); for a locally exported ticket (.kirbi) it is commonly stored
// unencrypted with etype 0.
type EncKrbCredPart struct {
	TicketInfo []KrbCredInfo
	Nonce      int
	Timestamp  time.Time
	Usec       int
}

type encKrbCredPartInner struct {
	TicketInfo []krbCredInfoInner `asn1:"explicit,tag:0"`
	Nonce      int                `asn1:"explicit,tag:1,optional"`
	Timestamp  time.Time          `asn1:"explicit,tag:2,optional,generalized"`
	Usec       int                `asn1:"explicit,tag:3,optional"`
	SAddress   HostAddress        `asn1:"explicit,tag:4,optional"`
	RAddress   HostAddress        `asn1:"explicit,tag:5,optional"`
}

// Marshal encodes the EncKrbCredPart as an ASN.1 APPLICATION[29] SEQUENCE.
func (e *EncKrbCredPart) Marshal() ([]byte, error) {
	var infoTLVs []byte
	for i := range e.TicketInfo {
		b, err := e.TicketInfo[i].marshal()
		if err != nil {
			return nil, fmt.Errorf("enckrbcredpart: ticket-info[%d]: %w", i, err)
		}
		infoTLVs = append(infoTLVs, b...)
	}
	infoSeq, err := asn1.Marshal(asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSequence, IsCompound: true, Bytes: infoTLVs})
	if err != nil {
		return nil, err
	}
	ticketInfoElem, err := derExplicit(0, infoSeq)
	if err != nil {
		return nil, err
	}

	elems := [][]byte{ticketInfoElem}
	if e.Nonce != 0 {
		n, err := asn1.MarshalWithParams(e.Nonce, "explicit,tag:1")
		if err != nil {
			return nil, err
		}
		elems = append(elems, n)
	}
	if !e.Timestamp.IsZero() {
		ts, err := asn1.MarshalWithParams(normalizeTime(e.Timestamp), "explicit,tag:2,generalized")
		if err != nil {
			return nil, err
		}
		elems = append(elems, ts)
		u, err := asn1.MarshalWithParams(e.Usec, "explicit,tag:3")
		if err != nil {
			return nil, err
		}
		elems = append(elems, u)
	}

	seq, err := derSequence(elems...)
	if err != nil {
		return nil, err
	}
	return wrapApplication(29, seq)
}

// Unmarshal decodes an EncKrbCredPart from an APPLICATION[29] SEQUENCE.
func (e *EncKrbCredPart) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, 29)
	if err != nil {
		return 0, fmt.Errorf("enckrbcredpart: %w", err)
	}
	var inner encKrbCredPartInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("enckrbcredpart inner unmarshal: %w", err)
	}
	e.TicketInfo = make([]KrbCredInfo, len(inner.TicketInfo))
	for i := range inner.TicketInfo {
		e.TicketInfo[i] = inner.TicketInfo[i].toPublic()
	}
	e.Nonce = inner.Nonce
	e.Timestamp = inner.Timestamp
	e.Usec = inner.Usec
	return consumed, nil
}

// KRBCred is a Kerberos KRB-CRED message (APPLICATION[22]), RFC 4120
// Section 5.8.1.
type KRBCred struct {
	PVNO    int
	MsgType int
	// Tickets holds the parsed tickets.
	Tickets []Ticket
	// TicketsRaw holds the raw APPLICATION[1] bytes of each ticket, preferred on
	// Marshal to re-emit exactly what a KDC issued.
	TicketsRaw [][]byte
	// EncPart is the (usually unencrypted, etype 0) EncKrbCredPart.
	EncPart EncryptedData
}

type krbCredInner struct {
	PVNO    int             `asn1:"explicit,tag:0"`
	MsgType int             `asn1:"explicit,tag:1"`
	Tickets []asn1.RawValue `asn1:"explicit,tag:2"`
	EncPart EncryptedData   `asn1:"explicit,tag:3"`
}

// Marshal encodes the KRB-CRED as an ASN.1 APPLICATION[22] SEQUENCE.
func (c *KRBCred) Marshal() ([]byte, error) {
	pvno, err := asn1.MarshalWithParams(KerberosV5, "explicit,tag:0")
	if err != nil {
		return nil, err
	}
	msgType, err := asn1.MarshalWithParams(MsgTypeKRBCred, "explicit,tag:1")
	if err != nil {
		return nil, err
	}

	// tickets [2] SEQUENCE OF Ticket — each Ticket is its own APPLICATION[1] TLV.
	var tktTLVs []byte
	if len(c.TicketsRaw) > 0 {
		for _, raw := range c.TicketsRaw {
			tktTLVs = append(tktTLVs, raw...)
		}
	} else {
		for i := range c.Tickets {
			b, err := c.Tickets[i].Marshal()
			if err != nil {
				return nil, fmt.Errorf("krbcred: ticket[%d]: %w", i, err)
			}
			tktTLVs = append(tktTLVs, b...)
		}
	}
	tktSeq, err := asn1.Marshal(asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSequence, IsCompound: true, Bytes: tktTLVs})
	if err != nil {
		return nil, err
	}
	ticketsElem, err := derExplicit(2, tktSeq)
	if err != nil {
		return nil, err
	}

	encPart, err := asn1.MarshalWithParams(c.EncPart, "explicit,tag:3")
	if err != nil {
		return nil, err
	}

	seq, err := derSequence(pvno, msgType, ticketsElem, encPart)
	if err != nil {
		return nil, err
	}
	return wrapApplication(MsgTypeKRBCred, seq)
}

// Unmarshal decodes a KRB-CRED from an APPLICATION[22] SEQUENCE.
func (c *KRBCred) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, MsgTypeKRBCred)
	if err != nil {
		return 0, fmt.Errorf("krbcred: %w", err)
	}
	var inner krbCredInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("krbcred inner unmarshal: %w", err)
	}
	if err := validateMessageHeader("krbcred", inner.PVNO, inner.MsgType, MsgTypeKRBCred); err != nil {
		return 0, err
	}
	c.PVNO = inner.PVNO
	c.MsgType = inner.MsgType
	c.EncPart = inner.EncPart
	c.Tickets = make([]Ticket, 0, len(inner.Tickets))
	c.TicketsRaw = make([][]byte, 0, len(inner.Tickets))
	for i := range inner.Tickets {
		raw := inner.Tickets[i].FullBytes
		var t Ticket
		if _, err := t.Unmarshal(raw); err != nil {
			return 0, fmt.Errorf("krbcred ticket[%d]: %w", i, err)
		}
		c.Tickets = append(c.Tickets, t)
		c.TicketsRaw = append(c.TicketsRaw, raw)
	}
	return consumed, nil
}
