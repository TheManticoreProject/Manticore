package messages

import (
	"encoding/asn1"
	"fmt"
	"time"
)

// KRBApplicationBody is the application-data body shared by KRB-SAFE and the
// encrypted part of KRB-PRIV (RFC 4120 sections 5.6.1 and 5.7.1).
type KRBApplicationBody struct {
	UserData         []byte      `asn1:"explicit,tag:0"`
	Timestamp        time.Time   `asn1:"explicit,tag:1,optional,generalized"`
	Usec             int         `asn1:"explicit,tag:2,optional"`
	SequenceNumber   int         `asn1:"explicit,tag:3,optional"`
	SenderAddress    HostAddress `asn1:"explicit,tag:4"`
	RecipientAddress HostAddress `asn1:"explicit,tag:5,optional"`
}

func (b KRBApplicationBody) Marshal() ([]byte, error) {
	if !b.Timestamp.IsZero() {
		b.Timestamp = normalizeTime(b.Timestamp)
	}
	return asn1.Marshal(b)
}

func (b *KRBApplicationBody) Unmarshal(data []byte) (int, error) {
	rest, err := asn1.Unmarshal(data, b)
	if err != nil {
		return 0, err
	}
	return len(data) - len(rest), nil
}

type krbSafeInner struct {
	PVNO     int                `asn1:"explicit,tag:0"`
	MsgType  int                `asn1:"explicit,tag:1"`
	SafeBody KRBApplicationBody `asn1:"explicit,tag:2"`
	Checksum Checksum           `asn1:"explicit,tag:3"`
}

// KRBSafe is an integrity-protected KRB-SAFE message (APPLICATION 20).
type KRBSafe struct {
	PVNO, MsgType int
	SafeBody      KRBApplicationBody
	Checksum      Checksum
}

func (m *KRBSafe) Marshal() ([]byte, error) {
	seq, err := asn1.Marshal(krbSafeInner{KerberosV5, MsgTypeKRBSafe, m.SafeBody, m.Checksum})
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
	m.PVNO, m.MsgType, m.SafeBody, m.Checksum = inner.PVNO, inner.MsgType, inner.SafeBody, inner.Checksum
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
