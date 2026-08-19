package messages

import (
	"encoding/asn1"
	"fmt"
	"time"
)

// apRepInner is the inner SEQUENCE of an AP-REP message.
type apRepInner struct {
	PVNO    int           `asn1:"explicit,tag:0"`
	MsgType int           `asn1:"explicit,tag:1"`
	EncPart EncryptedData `asn1:"explicit,tag:2"`
}

// APRep is a Kerberos AP-REP (Application Reply) message, APPLICATION[15], as
// defined in RFC 4120 Section 5.5.2. A service returns it to the client only
// when the AP-REQ set the mutual-required option; it proves the service holds
// the ticket session key. The enc-part is an encrypted EncAPRepPart (key
// usage 12, under the ticket session key).
type APRep struct {
	PVNO    int
	MsgType int
	EncPart EncryptedData
}

// Marshal encodes the AP-REP as an ASN.1 APPLICATION[15] wrapped SEQUENCE.
func (r *APRep) Marshal() ([]byte, error) {
	inner := apRepInner{
		PVNO:    KerberosV5,
		MsgType: MsgTypeAPRep,
		EncPart: r.EncPart,
	}
	seq_bytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(MsgTypeAPRep, seq_bytes)
}

// Unmarshal decodes an AP-REP from an ASN.1 APPLICATION[15] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (r *APRep) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, MsgTypeAPRep)
	if err != nil {
		return 0, fmt.Errorf("aprep: %w", err)
	}
	var inner apRepInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("aprep inner unmarshal: %w", err)
	}
	if err := validateMessageHeader("aprep", inner.PVNO, inner.MsgType, MsgTypeAPRep); err != nil {
		return 0, err
	}
	r.PVNO = inner.PVNO
	r.MsgType = inner.MsgType
	r.EncPart = inner.EncPart
	return consumed, nil
}

// encAPRepPartInner is the inner SEQUENCE of an EncAPRepPart. All fields are
// numeric/key material — no GeneralString — so a plain struct suffices.
type encAPRepPartInner struct {
	CTime     time.Time     `asn1:"explicit,tag:0,generalized"`
	CUSec     int           `asn1:"explicit,tag:1"`
	SubKey    EncryptionKey `asn1:"explicit,tag:2,optional"`
	SeqNumber int           `asn1:"explicit,tag:3,optional"`
}

// EncAPRepPart is the decrypted enc-part of an AP-REP (APPLICATION[27]), as
// defined in RFC 4120 Section 5.5.2. The client verifies that CTime/CUSec echo
// the values it placed in its Authenticator, confirming the service decrypted
// the ticket and thus holds the session key (mutual authentication).
type EncAPRepPart struct {
	// CTime must echo the ctime from the client's Authenticator.
	CTime time.Time
	// CUSec must echo the cusec from the client's Authenticator.
	CUSec int
	// SubKey is an optional service-chosen sub-session key.
	SubKey *EncryptionKey
	// SeqNumber is the optional service sequence number.
	SeqNumber int
}

// Marshal encodes the EncAPRepPart as an ASN.1 APPLICATION[27] wrapped SEQUENCE.
func (e *EncAPRepPart) Marshal() ([]byte, error) {
	inner := encAPRepPartInner{
		CTime:     normalizeTime(e.CTime),
		CUSec:     e.CUSec,
		SeqNumber: e.SeqNumber,
	}
	if e.SubKey != nil {
		inner.SubKey = *e.SubKey
	}
	seq_bytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(27, seq_bytes)
}

// Unmarshal decodes an EncAPRepPart from an ASN.1 APPLICATION[27] wrapped
// SEQUENCE. Returns the number of bytes consumed from data.
func (e *EncAPRepPart) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, 27)
	if err != nil {
		return 0, fmt.Errorf("encapreppart: %w", err)
	}
	var inner encAPRepPartInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("encapreppart inner unmarshal: %w", err)
	}
	e.CTime = inner.CTime
	e.CUSec = inner.CUSec
	e.SeqNumber = inner.SeqNumber
	if inner.SubKey.KeyType != 0 {
		sk := inner.SubKey
		e.SubKey = &sk
	}
	return consumed, nil
}
