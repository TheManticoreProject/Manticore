package messages

import (
	"encoding/asn1"
	"fmt"
	"time"
)

// krbErrorInner is the inner SEQUENCE of a KRB-ERROR message.
type krbErrorInner struct {
	// PVNO is the Kerberos protocol version (always 5).
	PVNO int `asn1:"explicit,tag:0"`
	// MsgType is the message type (always MsgTypeError = 30).
	MsgType int `asn1:"explicit,tag:1"`
	// CTime is the client time when the error occurred (optional).
	CTime time.Time `asn1:"explicit,tag:2,optional,generalized"`
	// CUSec is the microseconds component of CTime (optional).
	CUSec int `asn1:"explicit,tag:3,optional"`
	// STime is the server time when the error occurred.
	STime time.Time `asn1:"explicit,tag:4,generalized"`
	// SUSec is the microseconds component of STime.
	SUSec int `asn1:"explicit,tag:5"`
	// ErrorCode is the Kerberos error code.
	ErrorCode int `asn1:"explicit,tag:6"`
	// CRealm is the client's realm (optional).
	CRealm string `asn1:"explicit,tag:7,optional,generalstring"`
	// CName is the client's principal name (optional).
	CName PrincipalName `asn1:"explicit,tag:8,optional"`
	// Realm is the server's realm.
	Realm string `asn1:"explicit,tag:9,generalstring"`
	// SName is the server's principal name.
	SName PrincipalName `asn1:"explicit,tag:10"`
	// EText is an optional error text string.
	EText string `asn1:"explicit,tag:11,optional,utf8"`
	// EData contains additional error data (optional, e.g. PA-ETYPE-INFO2).
	EData []byte `asn1:"explicit,tag:12,optional"`
}

// KRBError is a Kerberos KRB-ERROR message (APPLICATION[30]),
// as defined in RFC 4120 Section 5.9.1.
// It is sent by the KDC when an error occurs processing a request.
type KRBError struct {
	// PVNO is the Kerberos protocol version.
	PVNO int
	// MsgType is the message type (MsgTypeError = 30).
	MsgType int
	// STime is the server time at which the error occurred.
	STime time.Time
	// SUSec is the microsecond component of STime.
	SUSec int
	// ErrorCode identifies the specific error.
	ErrorCode int
	// Realm is the server's realm.
	Realm string
	// SName is the server's principal name.
	SName PrincipalName
	// EText is a human-readable error description.
	EText string
	// EData contains additional structured error information.
	EData []byte
}

// Error implements the error interface, returning a description of the KRB error.
func (e *KRBError) Error() string {
	return fmt.Sprintf("KRB Error %d: %s", e.ErrorCode, e.EText)
}

// Marshal encodes the KRBError as an ASN.1 APPLICATION[30] wrapped SEQUENCE.
func (e *KRBError) Marshal() ([]byte, error) {
	inner := krbErrorInner{
		PVNO:      KerberosV5,
		MsgType:   MsgTypeError,
		STime:     e.STime,
		SUSec:     e.SUSec,
		ErrorCode: e.ErrorCode,
		Realm:     e.Realm,
		SName:     e.SName,
		EText:     e.EText,
		EData:     e.EData,
	}
	seq_bytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(MsgTypeError, seq_bytes)
}

// Unmarshal decodes a KRBError from an ASN.1 APPLICATION[30] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (e *KRBError) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, MsgTypeError)
	if err != nil {
		return 0, fmt.Errorf("krberror: %w", err)
	}

	var inner krbErrorInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("krberror inner unmarshal: %w", err)
	}

	e.PVNO = inner.PVNO
	e.MsgType = inner.MsgType
	e.STime = inner.STime
	e.SUSec = inner.SUSec
	e.ErrorCode = inner.ErrorCode
	e.Realm = inner.Realm
	e.SName = inner.SName
	e.EText = inner.EText
	e.EData = inner.EData
	return consumed, nil
}
