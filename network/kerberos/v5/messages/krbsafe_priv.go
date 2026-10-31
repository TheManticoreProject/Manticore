package messages

import (
	"encoding/asn1"
	"fmt"
	"math"
	"time"
)

// KRBApplicationBody is the application-data body shared by KRB-SAFE and the
// encrypted part of KRB-PRIV (RFC 4120 sections 5.6.1 and 5.7.1). Usec and
// SequenceNumber are optional: nil omits the field, and a non-nil pointer is
// encoded even when it points at 0.
type KRBApplicationBody struct {
	UserData         []byte
	Timestamp        time.Time
	Usec             *int
	SequenceNumber   *int
	SenderAddress    HostAddress
	RecipientAddress HostAddress
}

// krbApplicationBodyAbsent marks an absent usec / seq-number in the wire form.
// It lies outside every valid value (Microseconds and UInt32), so 0 is encoded.
const krbApplicationBodyAbsent = -1 << 62

// krbApplicationBodyWire is the ASN.1 form of KRBApplicationBody.
type krbApplicationBodyWire struct {
	UserData         []byte      `asn1:"explicit,tag:0"`
	Timestamp        time.Time   `asn1:"explicit,tag:1,optional,generalized"`
	Usec             int64       `asn1:"explicit,tag:2,optional,default:-4611686018427387904"`
	SequenceNumber   int64       `asn1:"explicit,tag:3,optional,default:-4611686018427387904"`
	SenderAddress    HostAddress `asn1:"explicit,tag:4"`
	RecipientAddress HostAddress `asn1:"explicit,tag:5,optional"`
}

func (b KRBApplicationBody) Marshal() ([]byte, error) {
	w := krbApplicationBodyWire{
		UserData: b.UserData, Timestamp: b.Timestamp,
		Usec: krbApplicationBodyAbsent, SequenceNumber: krbApplicationBodyAbsent,
		SenderAddress: b.SenderAddress, RecipientAddress: b.RecipientAddress,
	}
	if !w.Timestamp.IsZero() {
		w.Timestamp = normalizeTime(w.Timestamp)
	}
	if b.Usec != nil {
		w.Usec = int64(*b.Usec)
	}
	if b.SequenceNumber != nil {
		w.SequenceNumber = int64(*b.SequenceNumber)
	}
	return asn1.Marshal(w)
}

func (b *KRBApplicationBody) Unmarshal(data []byte) (int, error) {
	var w krbApplicationBodyWire
	rest, err := asn1.Unmarshal(data, &w)
	if err != nil {
		return 0, err
	}
	*b = KRBApplicationBody{
		UserData: w.UserData, Timestamp: w.Timestamp,
		SenderAddress: w.SenderAddress, RecipientAddress: w.RecipientAddress,
	}
	// usec is Microseconds (0..999999) and seq-number is UInt32 (RFC 4120
	// section 5.2.4). Out-of-range values are refused rather than converted,
	// which on a 32-bit int would silently wrap a sequence number.
	if w.Usec != krbApplicationBodyAbsent {
		if w.Usec < 0 || w.Usec > 999999 {
			return 0, fmt.Errorf("usec %d is outside 0..999999", w.Usec)
		}
		usec := int(w.Usec)
		b.Usec = &usec
	}
	if w.SequenceNumber != krbApplicationBodyAbsent {
		if w.SequenceNumber < 0 || w.SequenceNumber > math.MaxUint32 {
			return 0, fmt.Errorf("seq-number %d is outside the UInt32 range", w.SequenceNumber)
		}
		seq := int(w.SequenceNumber)
		b.SequenceNumber = &seq
	}
	return len(data) - len(rest), nil
}

// krbSafeInner carries safe-body as an untagged RawValue so its received bytes
// are kept verbatim; the [2] wrapper is checked by KRBSafe.Unmarshal.
type krbSafeInner struct {
	PVNO     int `asn1:"explicit,tag:0"`
	MsgType  int `asn1:"explicit,tag:1"`
	SafeBody asn1.RawValue
	Checksum Checksum `asn1:"explicit,tag:3"`
}

// KRBSafe is an integrity-protected KRB-SAFE message (APPLICATION 20).
type KRBSafe struct {
	PVNO, MsgType int
	SafeBody      KRBApplicationBody
	// SafeBodyRaw holds the safe-body DER exactly as received. The KRB-SAFE
	// checksum covers these bytes (RFC 4120 section 5.6.1), so verify against
	// them rather than a re-encoding of SafeBody. Set by Unmarshal.
	SafeBodyRaw []byte
	Checksum    Checksum
}

func (m *KRBSafe) Marshal() ([]byte, error) {
	body, err := m.SafeBody.Marshal()
	if err != nil {
		return nil, err
	}
	safeBody := asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 2, IsCompound: true, Bytes: body}
	seq, err := asn1.Marshal(krbSafeInner{KerberosV5, MsgTypeKRBSafe, safeBody, m.Checksum})
	if err != nil {
		return nil, err
	}
	return wrapApplication(MsgTypeKRBSafe, seq)
}

func (m *KRBSafe) Unmarshal(data []byte) (int, error) {
	innerBytes, consumed, err := unwrapApplication(data, MsgTypeKRBSafe)
	if err != nil {
		return 0, fmt.Errorf("krb-safe: %w", err)
	}
	var inner krbSafeInner
	if _, err := asn1.Unmarshal(innerBytes, &inner); err != nil {
		return 0, fmt.Errorf("krb-safe inner: %w", err)
	}
	if err := validateMessageHeader("krb-safe", inner.PVNO, inner.MsgType, MsgTypeKRBSafe); err != nil {
		return 0, err
	}
	if inner.SafeBody.Class != asn1.ClassContextSpecific || inner.SafeBody.Tag != 2 || !inner.SafeBody.IsCompound {
		return 0, fmt.Errorf("krb-safe: safe-body is not [2]")
	}
	var body KRBApplicationBody
	n, err := body.Unmarshal(inner.SafeBody.Bytes)
	if err != nil {
		return 0, fmt.Errorf("krb-safe safe-body: %w", err)
	}
	if n != len(inner.SafeBody.Bytes) {
		return 0, fmt.Errorf("krb-safe: trailing data after safe-body")
	}
	m.PVNO, m.MsgType, m.SafeBody, m.Checksum = inner.PVNO, inner.MsgType, body, inner.Checksum
	m.SafeBodyRaw = append([]byte(nil), inner.SafeBody.Bytes...)
	return consumed, nil
}

type krbPrivInner struct {
	PVNO    int           `asn1:"explicit,tag:0"`
	MsgType int           `asn1:"explicit,tag:1"`
	EncPart EncryptedData `asn1:"explicit,tag:3"`
}

// KRBPriv is a confidentiality-protected KRB-PRIV message (APPLICATION 21).
type KRBPriv struct {
	PVNO, MsgType int
	EncPart       EncryptedData
}

func (m *KRBPriv) Marshal() ([]byte, error) {
	seq, err := asn1.Marshal(krbPrivInner{KerberosV5, MsgTypeKRBPriv, m.EncPart})
	if err != nil {
		return nil, err
	}
	return wrapApplication(MsgTypeKRBPriv, seq)
}

func (m *KRBPriv) Unmarshal(data []byte) (int, error) {
	innerBytes, consumed, err := unwrapApplication(data, MsgTypeKRBPriv)
	if err != nil {
		return 0, fmt.Errorf("krb-priv: %w", err)
	}
	var inner krbPrivInner
	if _, err := asn1.Unmarshal(innerBytes, &inner); err != nil {
		return 0, fmt.Errorf("krb-priv inner: %w", err)
	}
	if err := validateMessageHeader("krb-priv", inner.PVNO, inner.MsgType, MsgTypeKRBPriv); err != nil {
		return 0, err
	}
	m.PVNO, m.MsgType, m.EncPart = inner.PVNO, inner.MsgType, inner.EncPart
	return consumed, nil
}

// EncKRBPrivPart is the decrypted KRB-PRIV enc-part (APPLICATION 28).
type EncKRBPrivPart KRBApplicationBody

func (p *EncKRBPrivPart) Marshal() ([]byte, error) {
	body, err := KRBApplicationBody(*p).Marshal()
	if err != nil {
		return nil, err
	}
	return wrapApplication(28, body)
}

func (p *EncKRBPrivPart) Unmarshal(data []byte) (int, error) {
	body, consumed, err := unwrapApplication(data, 28)
	if err != nil {
		return 0, fmt.Errorf("enc-krb-priv-part: %w", err)
	}
	var value KRBApplicationBody
	if _, err := value.Unmarshal(body); err != nil {
		return 0, err
	}
	*p = EncKRBPrivPart(value)
	return consumed, nil
}
