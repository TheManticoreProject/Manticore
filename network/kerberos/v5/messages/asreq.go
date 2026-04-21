package messages

import (
	"encoding/asn1"
	"fmt"
)

// asReqInner is the inner SEQUENCE of an AS-REQ message for unmarshaling.
// It is wrapped in an APPLICATION[10] tag by Marshal.
type asReqInner struct {
	// PVNO is the Kerberos protocol version number (always 5).
	PVNO int `asn1:"explicit,tag:1"`
	// MsgType is the message type (always MsgTypeASReq = 10).
	MsgType int `asn1:"explicit,tag:2"`
	// PAData contains optional pre-authentication data.
	PAData []PAData `asn1:"explicit,tag:3,optional"`
	// ReqBody is the KDC request body.
	ReqBody KDCReqBody `asn1:"explicit,tag:4"`
}

// asReqMarshal is the inner SEQUENCE for marshaling — uses GeneralString types.
type asReqMarshal struct {
	PVNO    int              `asn1:"explicit,tag:1"`
	MsgType int              `asn1:"explicit,tag:2"`
	PAData  []PAData         `asn1:"explicit,tag:3,optional"`
	ReqBody kdcReqBodyMarshal `asn1:"explicit,tag:4"`
}

// ASReq is a Kerberos AS-REQ (Authentication Service Request) message,
// APPLICATION[10], as defined in RFC 4120 Section 5.4.1.
// It is sent by the client to the KDC to request a TGT.
type ASReq struct {
	// PVNO is the Kerberos protocol version (always 5).
	PVNO int
	// MsgType is the message type (always MsgTypeASReq = 10).
	MsgType int
	// PAData contains pre-authentication data (e.g. PA-ENC-TIMESTAMP).
	PAData []PAData
	// ReqBody is the KDC request body containing client/server names and options.
	ReqBody KDCReqBody
}

// Marshal encodes the AS-REQ as an ASN.1 APPLICATION[10] wrapped SEQUENCE.
func (r *ASReq) Marshal() ([]byte, error) {
	inner := asReqMarshal{
		PVNO:    KerberosV5,
		MsgType: MsgTypeASReq,
		PAData:  r.PAData,
		ReqBody: marshalKDCReqBody(r.ReqBody),
	}
	seq_bytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(MsgTypeASReq, seq_bytes)
}

// Unmarshal decodes an AS-REQ from an ASN.1 APPLICATION[10] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (r *ASReq) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, MsgTypeASReq)
	if err != nil {
		return 0, fmt.Errorf("asreq: %w", err)
	}

	var inner asReqInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("asreq inner unmarshal: %w", err)
	}

	r.PVNO = inner.PVNO
	r.MsgType = inner.MsgType
	r.PAData = inner.PAData
	r.ReqBody = inner.ReqBody
	return consumed, nil
}
