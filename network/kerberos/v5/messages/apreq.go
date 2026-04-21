package messages

import (
	"encoding/asn1"
	"fmt"
)

// apReqInner is the inner SEQUENCE of an AP-REQ message.
type apReqInner struct {
	// PVNO is the Kerberos protocol version (always 5).
	PVNO int `asn1:"explicit,tag:0"`
	// MsgType is the message type (always MsgTypeAPReq = 14).
	MsgType int `asn1:"explicit,tag:1"`
	// APOptions contains bit flags for the AP request.
	APOptions asn1.BitString `asn1:"explicit,tag:2"`
	// Ticket is pre-encoded as [3] EXPLICIT { APPLICATION[1] bytes }.
	// Go ignores explicit,tag:N for asn1.RawValue with FullBytes, so we pre-build
	// the context wrapper and store it here without a struct tag.
	Ticket asn1.RawValue
	// Authenticator is the encrypted authenticator.
	Authenticator EncryptedData `asn1:"explicit,tag:4"`
}

// APReq is a Kerberos AP-REQ (Application Request) message,
// APPLICATION[14], as defined in RFC 4120 Section 5.5.1.
// It is sent by the client to a service as part of mutual authentication,
// and is also embedded in TGS-REQ PA-DATA (PA-TGS-REQ).
type APReq struct {
	// PVNO is the Kerberos protocol version (always 5).
	PVNO int
	// MsgType is the message type (always MsgTypeAPReq = 14).
	MsgType int
	// APOptions contains bit flags controlling the AP exchange.
	APOptions asn1.BitString
	// Ticket is the service ticket (parsed form).
	Ticket Ticket
	// TicketRaw holds raw APPLICATION[1] bytes from the KDC, used verbatim in Marshal
	// to avoid re-encoding the ticket (which might differ from the KDC's original encoding).
	TicketRaw []byte
	// Authenticator is the encrypted Authenticator proving the client's identity.
	Authenticator EncryptedData
}

// Marshal encodes the AP-REQ as an ASN.1 APPLICATION[14] wrapped SEQUENCE.
func (r *APReq) Marshal() ([]byte, error) {
	// Prefer raw bytes from KDC (TicketRaw) to avoid re-encoding differences.
	tkt_bytes := r.TicketRaw
	if len(tkt_bytes) == 0 {
		var err error
		tkt_bytes, err = r.Ticket.Marshal()
		if err != nil {
			return nil, err
		}
	}
	// Pre-encode [3] EXPLICIT { APPLICATION[1] bytes }.
	// Go ignores explicit,tag:N for asn1.RawValue with FullBytes set, so we build
	// the [3] wrapper manually using Bytes (which Go wraps with Class/Tag/IsCompound).
	tkt_raw := asn1.RawValue{
		Class:      asn1.ClassContextSpecific,
		Tag:        3,
		IsCompound: true,
		Bytes:      tkt_bytes,
	}

	inner := apReqInner{
		PVNO:          KerberosV5,
		MsgType:       MsgTypeAPReq,
		APOptions:     r.APOptions,
		Ticket:        tkt_raw,
		Authenticator: r.Authenticator,
	}
	seq_bytes, err := asn1.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return wrapApplication(MsgTypeAPReq, seq_bytes)
}

// Unmarshal decodes an AP-REQ from an ASN.1 APPLICATION[14] wrapped SEQUENCE.
// Returns the number of bytes consumed from data.
func (r *APReq) Unmarshal(data []byte) (int, error) {
	inner_bytes, consumed, err := unwrapApplication(data, MsgTypeAPReq)
	if err != nil {
		return 0, fmt.Errorf("apreq: %w", err)
	}

	var inner apReqInner
	if _, err := asn1.Unmarshal(inner_bytes, &inner); err != nil {
		return 0, fmt.Errorf("apreq inner unmarshal: %w", err)
	}

	r.PVNO = inner.PVNO
	r.MsgType = inner.MsgType
	r.APOptions = inner.APOptions
	r.Authenticator = inner.Authenticator

	// Unmarshal the ticket from the raw value
	tkt_raw_bytes, err := asn1.Marshal(inner.Ticket)
	if err != nil {
		return 0, err
	}
	if _, err := r.Ticket.Unmarshal(tkt_raw_bytes); err != nil {
		return 0, fmt.Errorf("apreq ticket unmarshal: %w", err)
	}

	return consumed, nil
}
