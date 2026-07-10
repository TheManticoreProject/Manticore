package messages

import (
	"encoding/asn1"
	"fmt"
)

// tgsReqInner is the inner SEQUENCE of a TGS-REQ message for unmarshaling.
type tgsReqInner struct {
	// PVNO is the Kerberos protocol version (always 5).
	PVNO int `asn1:"explicit,tag:1"`
	// MsgType is the message type (always MsgTypeTGSReq = 12).
	MsgType int `asn1:"explicit,tag:2"`
	// PAData contains pre-authentication data (must include PA-TGS-REQ with AP-REQ).
	PAData []PAData `asn1:"explicit,tag:3,optional"`
	// ReqBody is the KDC request body specifying the desired service ticket. It is
	// decoded through kdcReqBodyParse (see fast.go) because the canonical
	// KDCReqBody does not round-trip through Go's encoding/asn1 on decode.
	ReqBody kdcReqBodyParse `asn1:"explicit,tag:4"`
}

// tgsReqMarshal is the inner SEQUENCE for marshaling — uses GeneralString types.
// ReqBody is a pre-encoded [4] EXPLICIT { KDC-REQ-BODY } RawValue so that
// additional-tickets can be spliced correctly (see encodeKDCReqBodyForTGS).
type tgsReqMarshal struct {
	PVNO    int           `asn1:"explicit,tag:1"`
	MsgType int           `asn1:"explicit,tag:2"`
	PAData  []PAData      `asn1:"explicit,tag:3,optional"`
	ReqBody asn1.RawValue // pre-encoded [4] EXPLICIT { KDC-REQ-BODY }
}

// TGSReq is a Kerberos TGS-REQ (Ticket Granting Service Request) message,
// APPLICATION[12], as defined in RFC 4120 Section 5.4.1.
// It is sent by the client to the TGS to request a service ticket.
// The PA-TGS-REQ pre-authentication data must contain an AP-REQ with the TGT.
type TGSReq struct {
	// PVNO is the Kerberos protocol version (always 5).
	PVNO int
	// MsgType is the message type (always MsgTypeTGSReq = 12).
	MsgType int
	// PAData contains the PA-TGS-REQ with the AP-REQ carrying the TGT.
	PAData []PAData
	// ReqBody is the request body specifying the requested service ticket parameters.
	ReqBody KDCReqBody
}

// Marshal encodes the TGS-REQ as an ASN.1 APPLICATION[12] wrapped SEQUENCE.
func (r *TGSReq) Marshal() ([]byte, error) {
	bodyTLV, err := encodeKDCReqBodyForTGS(r.ReqBody)
	if err != nil {
		return nil, err
	}
	inner := tgsReqMarshal{
		PVNO:    KerberosV5,
		MsgType: MsgTypeTGSReq,
		PAData:  r.PAData,
		ReqBody: asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 4, IsCompound: true, Bytes: bodyTLV},
	}
	seq_bytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(MsgTypeTGSReq, seq_bytes)
}

// Unmarshal decodes a TGS-REQ from an ASN.1 APPLICATION[12] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (r *TGSReq) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, MsgTypeTGSReq)
	if err != nil {
		return 0, fmt.Errorf("tgsreq: %w", err)
	}

	var inner tgsReqInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("tgsreq inner unmarshal: %w", err)
	}

	r.PVNO = inner.PVNO
	r.MsgType = inner.MsgType
	r.PAData = inner.PAData
	r.ReqBody = inner.ReqBody.toKDCReqBody()
	return consumed, nil
}
