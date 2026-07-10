package messages

import (
	"encoding/asn1"
	"fmt"
	"time"
)

// This file implements the wire structures of RFC 6113 (A Generalized Framework
// for Kerberos Pre-Authentication — "FAST", Flexible Authentication Secure
// Tunneling), as profiled by MS-KILE for Kerberos armoring.
//
// FAST wraps an ordinary KDC-REQ inside an armored, integrity-protected
// envelope keyed by an "armor key". The client armors an AS-REQ/TGS-REQ with
// PA-FX-FAST (padata-type 136), carrying a KrbFastArmoredReq; the KDC replies
// with PA-FX-FAST carrying a KrbFastArmoredRep. Errors are delivered inside the
// armor as PA-FX-ERROR (137), and stateful replay protection cookies as
// PA-FX-COOKIE (133).
//
// The Kerberos ASN.1 modules use EXPLICIT tagging, so every context tag below
// is explicit, matching the rest of this package.

// FAST armor types (RFC 6113 §5.4.1). FX_FAST_ARMOR_AP_REQUEST carries a
// DER-encoded AP-REQ whose authenticator subkey, combined with the armor
// ticket's session key, yields the armor key.
const (
	FXFastArmorAPRequest = 1 // FX_FAST_ARMOR_AP_REQUEST
)

// FastOptions bit positions (RFC 6113 §5.4.2). Bit 0 is the MSB.
const (
	// FastOptionHideClientNames requests that the KDC omit the client identity
	// from error replies (a critical option).
	FastOptionHideClientNames = 1
)

// KrbFastArmor is the FAST armor descriptor (RFC 6113 §5.4.1):
//
//	KrbFastArmor ::= SEQUENCE {
//	    armor-type  [0] Int32,
//	    armor-value [1] OCTET STRING,
//	    ...
//	}
type KrbFastArmor struct {
	ArmorType  int    `asn1:"explicit,tag:0"`
	ArmorValue []byte `asn1:"explicit,tag:1"`
}

// Marshal encodes the KrbFastArmor SEQUENCE.
func (a *KrbFastArmor) Marshal() ([]byte, error) { return asn1.Marshal(*a) }

// Unmarshal decodes a KrbFastArmor SEQUENCE, returning bytes consumed.
func (a *KrbFastArmor) Unmarshal(data []byte) (int, error) {
	rest, err := asn1.Unmarshal(data, a)
	if err != nil {
		return 0, err
	}
	return len(data) - len(rest), nil
}

// krbFastReqMarshal is the GeneralString-aware marshal form of KrbFastReq: the
// req-body is a KDC-REQ-BODY whose realm/principal names must be GeneralString
// (Go's encoding/asn1 would otherwise emit PrintableString).
type krbFastReqMarshal struct {
	FastOptions asn1.BitString    `asn1:"explicit,tag:0"`
	PAData      []PAData          `asn1:"explicit,tag:1"`
	ReqBody     kdcReqBodyMarshal `asn1:"explicit,tag:2"`
}

// kdcReqBodyParse is a KDC-REQ-BODY unmarshal view. It exists because the
// canonical KDCReqBody carries AdditTickets/AdditTicketsRaw with asn1:"-", which
// Go's encoding/asn1 does not honor on decode (it would try to parse them as
// SEQUENCE fields and fail).
//
// The trailing optional fields enc-authorization-data [10] and additional-tickets
// [11] are decoded as asn1.RawValue rather than their concrete types: Go's
// encoding/asn1 raises "sequence truncated" when the final field of a SEQUENCE is
// an absent optional *value* struct (only pointers/slices/RawValue are skipped
// cleanly at end-of-sequence), and the [11] tickets are APPLICATION[1] TLVs the
// generic decoder cannot reconstruct. toKDCReqBody converts both.
type kdcReqBodyParse struct {
	KDCOptions   asn1.BitString `asn1:"explicit,tag:0"`
	CName        PrincipalName  `asn1:"explicit,tag:1,optional"`
	Realm        string         `asn1:"explicit,tag:2,generalstring"`
	SName        PrincipalName  `asn1:"explicit,tag:3,optional"`
	From         time.Time      `asn1:"explicit,tag:4,optional,generalized"`
	Till         time.Time      `asn1:"explicit,tag:5,generalized"`
	RTime        time.Time      `asn1:"explicit,tag:6,optional,generalized"`
	Nonce        int            `asn1:"explicit,tag:7"`
	EType        []int          `asn1:"explicit,tag:8"`
	Addresses    []HostAddress  `asn1:"explicit,tag:9,optional"`
	EncAuthData  asn1.RawValue  `asn1:"explicit,tag:10,optional"`
	AdditTickets asn1.RawValue  `asn1:"explicit,tag:11,optional"`
}

func (p kdcReqBodyParse) toKDCReqBody() KDCReqBody {
	b := KDCReqBody{
		KDCOptions: p.KDCOptions,
		CName:      p.CName,
		Realm:      p.Realm,
		SName:      p.SName,
		From:       p.From,
		Till:       p.Till,
		RTime:      p.RTime,
		Nonce:      p.Nonce,
		EType:      p.EType,
		Addresses:  p.Addresses,
	}
	// enc-authorization-data [10]: RawValue.Bytes is the inner EncryptedData TLV.
	if len(p.EncAuthData.Bytes) > 0 {
		var ed EncryptedData
		if _, err := asn1.Unmarshal(p.EncAuthData.Bytes, &ed); err == nil {
			b.EncAuthData = ed
		}
	}
	// additional-tickets [11]: RawValue.Bytes is the SEQUENCE OF Ticket TLV; each
	// element is an APPLICATION[1] ticket recovered verbatim into AdditTicketsRaw.
	if len(p.AdditTickets.Bytes) > 0 {
		var seq asn1.RawValue
		if _, err := asn1.Unmarshal(p.AdditTickets.Bytes, &seq); err == nil {
			rest := seq.Bytes
			for len(rest) > 0 {
				var tkt asn1.RawValue
				r, err := asn1.Unmarshal(rest, &tkt)
				if err != nil {
					break
				}
				b.AdditTicketsRaw = append(b.AdditTicketsRaw, tkt.FullBytes)
				rest = r
			}
		}
	}
	return b
}

// krbFastReqInner is the unmarshal form of KrbFastReq.
type krbFastReqInner struct {
	FastOptions asn1.BitString  `asn1:"explicit,tag:0"`
	PAData      []PAData        `asn1:"explicit,tag:1"`
	ReqBody     kdcReqBodyParse `asn1:"explicit,tag:2"`
}

// KrbFastReq is the plaintext of the enc-fast-req field (RFC 6113 §5.4.2):
//
//	KrbFastReq ::= SEQUENCE {
//	    fast-options [0] FastOptions,
//	    padata       [1] SEQUENCE OF PA-DATA,
//	    req-body     [2] KDC-REQ-BODY,
//	    ...
//	}
//
// The KDC uses this inner req-body and padata in preference to the outer,
// unprotected KDC-REQ.
type KrbFastReq struct {
	FastOptions asn1.BitString
	PAData      []PAData
	ReqBody     KDCReqBody
}

// Marshal encodes the KrbFastReq SEQUENCE with GeneralString-encoded names.
func (r *KrbFastReq) Marshal() ([]byte, error) {
	return asn1.Marshal(krbFastReqMarshal{
		FastOptions: r.FastOptions,
		PAData:      r.PAData,
		ReqBody:     marshalKDCReqBody(r.ReqBody),
	})
}

// Unmarshal decodes a KrbFastReq SEQUENCE, returning bytes consumed.
func (r *KrbFastReq) Unmarshal(data []byte) (int, error) {
	var inner krbFastReqInner
	rest, err := asn1.Unmarshal(data, &inner)
	if err != nil {
		return 0, err
	}
	r.FastOptions = inner.FastOptions
	r.PAData = inner.PAData
	r.ReqBody = inner.ReqBody.toKDCReqBody()
	return len(data) - len(rest), nil
}

// KrbFastArmoredReq is the armored FAST request (RFC 6113 §5.4.2):
//
//	KrbFastArmoredReq ::= SEQUENCE {
//	    armor        [0] KrbFastArmor OPTIONAL,
//	    req-checksum [1] Checksum,
//	    enc-fast-req [2] EncryptedData -- KrbFastReq --,
//	    ...
//	}
//
// For an AS-REQ the armor field MUST be present. req-checksum is a keyed
// checksum, computed with the armor key (key usage 50), over the outer
// KDC-REQ-BODY. enc-fast-req is the KrbFastReq encrypted under the armor key
// (key usage 51).
type KrbFastArmoredReq struct {
	Armor       *KrbFastArmor
	ReqChecksum Checksum
	EncFastReq  EncryptedData
}

type krbFastArmoredReqMarshal struct {
	Armor       KrbFastArmor  `asn1:"explicit,tag:0,optional"`
	ReqChecksum Checksum      `asn1:"explicit,tag:1"`
	EncFastReq  EncryptedData `asn1:"explicit,tag:2"`
}

type krbFastArmoredReqInner struct {
	Armor       KrbFastArmor  `asn1:"explicit,tag:0,optional"`
	ReqChecksum Checksum      `asn1:"explicit,tag:1"`
	EncFastReq  EncryptedData `asn1:"explicit,tag:2"`
}

// Marshal encodes the KrbFastArmoredReq SEQUENCE.
func (r *KrbFastArmoredReq) Marshal() ([]byte, error) {
	m := krbFastArmoredReqMarshal{ReqChecksum: r.ReqChecksum, EncFastReq: r.EncFastReq}
	if r.Armor != nil {
		m.Armor = *r.Armor
	}
	return asn1.Marshal(m)
}

// Unmarshal decodes a KrbFastArmoredReq SEQUENCE, returning bytes consumed.
func (r *KrbFastArmoredReq) Unmarshal(data []byte) (int, error) {
	var inner krbFastArmoredReqInner
	rest, err := asn1.Unmarshal(data, &inner)
	if err != nil {
		return 0, err
	}
	if inner.Armor.ArmorType != 0 || len(inner.Armor.ArmorValue) != 0 {
		a := inner.Armor
		r.Armor = &a
	}
	r.ReqChecksum = inner.ReqChecksum
	r.EncFastReq = inner.EncFastReq
	return len(data) - len(rest), nil
}

// KrbFastArmoredRep is the armored FAST reply (RFC 6113 §5.4.3):
//
//	KrbFastArmoredRep ::= SEQUENCE {
//	    enc-fast-rep [0] EncryptedData -- KrbFastResponse --,
//	    ...
//	}
//
// enc-fast-rep is a KrbFastResponse encrypted under the armor key (key usage 52).
type KrbFastArmoredRep struct {
	EncFastRep EncryptedData `asn1:"explicit,tag:0"`
}

// Marshal encodes the KrbFastArmoredRep SEQUENCE.
func (r *KrbFastArmoredRep) Marshal() ([]byte, error) { return asn1.Marshal(*r) }

// Unmarshal decodes a KrbFastArmoredRep SEQUENCE, returning bytes consumed.
func (r *KrbFastArmoredRep) Unmarshal(data []byte) (int, error) {
	rest, err := asn1.Unmarshal(data, r)
	if err != nil {
		return 0, err
	}
	return len(data) - len(rest), nil
}

// KrbFastResponse is the plaintext of enc-fast-rep (RFC 6113 §5.4.3):
//
//	KrbFastResponse ::= SEQUENCE {
//	    padata         [0] SEQUENCE OF PA-DATA,
//	    strengthen-key [1] EncryptionKey OPTIONAL,
//	    finished       [2] KrbFastFinished OPTIONAL,
//	    nonce          [3] UInt32,
//	    ...
//	}
//
// When strengthen-key is present the reply key is replaced by
// KRB-FX-CF2(strengthen-key, reply-key, "strengthenkey", "replykey"). nonce
// echoes the inner KDC-REQ-BODY nonce and MUST match it.
type KrbFastResponse struct {
	PAData        []PAData
	StrengthenKey *EncryptionKey
	Finished      *KrbFastFinished
	Nonce         int
}

type krbFastResponseMarshal struct {
	PAData        []PAData      `asn1:"explicit,tag:0"`
	StrengthenKey EncryptionKey `asn1:"explicit,tag:1,optional"`
	Finished      asn1.RawValue `asn1:"explicit,tag:2,optional"`
	Nonce         int           `asn1:"explicit,tag:3"`
}

type krbFastResponseInner struct {
	PAData        []PAData      `asn1:"explicit,tag:0"`
	StrengthenKey EncryptionKey `asn1:"explicit,tag:1,optional"`
	Finished      asn1.RawValue `asn1:"explicit,tag:2,optional"`
	Nonce         int           `asn1:"explicit,tag:3"`
}

// Marshal encodes the KrbFastResponse SEQUENCE.
func (r *KrbFastResponse) Marshal() ([]byte, error) {
	m := krbFastResponseMarshal{PAData: r.PAData, Nonce: r.Nonce}
	if r.StrengthenKey != nil {
		m.StrengthenKey = *r.StrengthenKey
	}
	if r.Finished != nil {
		fb, err := r.Finished.Marshal()
		if err != nil {
			return nil, err
		}
		m.Finished = asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 2, IsCompound: true, Bytes: fb}
	}
	return asn1.Marshal(m)
}

// Unmarshal decodes a KrbFastResponse SEQUENCE, returning bytes consumed.
func (r *KrbFastResponse) Unmarshal(data []byte) (int, error) {
	var inner krbFastResponseInner
	rest, err := asn1.Unmarshal(data, &inner)
	if err != nil {
		return 0, err
	}
	r.PAData = inner.PAData
	r.Nonce = inner.Nonce
	if inner.StrengthenKey.KeyType != 0 || len(inner.StrengthenKey.KeyValue) != 0 {
		sk := inner.StrengthenKey
		r.StrengthenKey = &sk
	}
	// Finished is [2] EXPLICIT { KrbFastFinished SEQUENCE }. Go leaves the outer
	// context tag on the RawValue and Bytes = the inner SEQUENCE TLV.
	if len(inner.Finished.Bytes) != 0 {
		var f KrbFastFinished
		if _, err := f.Unmarshal(inner.Finished.Bytes); err != nil {
			return 0, fmt.Errorf("krbfastresponse: finished: %w", err)
		}
		r.Finished = &f
	}
	return len(data) - len(rest), nil
}

// KrbFastFinished authenticates the FAST exchange to the client (RFC 6113
// §5.4.3):
//
//	KrbFastFinished ::= SEQUENCE {
//	    timestamp       [0] KerberosTime,
//	    usec            [1] Microseconds,
//	    crealm          [2] Realm,
//	    cname           [3] PrincipalName,
//	    ticket-checksum [4] Checksum,
//	    ...
//	}
//
// ticket-checksum is a keyed checksum over the issued ticket, computed with the
// armor key (key usage 53).
type KrbFastFinished struct {
	Timestamp      time.Time
	Usec           int
	CRealm         string
	CName          PrincipalName
	TicketChecksum Checksum
}

type krbFastFinishedMarshal struct {
	Timestamp      time.Time            `asn1:"explicit,tag:0,generalized"`
	Usec           int                  `asn1:"explicit,tag:1"`
	CRealm         asn1.RawValue        // pre-encoded [2] EXPLICIT { GeneralString }
	CName          PrincipalNameMarshal `asn1:"explicit,tag:3"`
	TicketChecksum Checksum             `asn1:"explicit,tag:4"`
}

type krbFastFinishedInner struct {
	Timestamp      time.Time     `asn1:"explicit,tag:0,generalized"`
	Usec           int           `asn1:"explicit,tag:1"`
	CRealm         string        `asn1:"explicit,tag:2,generalstring"`
	CName          PrincipalName `asn1:"explicit,tag:3"`
	TicketChecksum Checksum      `asn1:"explicit,tag:4"`
}

// Marshal encodes the KrbFastFinished SEQUENCE with GeneralString names.
func (f *KrbFastFinished) Marshal() ([]byte, error) {
	return asn1.Marshal(krbFastFinishedMarshal{
		Timestamp:      normalizeTime(f.Timestamp),
		Usec:           f.Usec,
		CRealm:         realmExplicit(2, f.CRealm),
		CName:          MarshalPrincipalName(f.CName),
		TicketChecksum: f.TicketChecksum,
	})
}

// Unmarshal decodes a KrbFastFinished SEQUENCE, returning bytes consumed.
func (f *KrbFastFinished) Unmarshal(data []byte) (int, error) {
	var inner krbFastFinishedInner
	rest, err := asn1.Unmarshal(data, &inner)
	if err != nil {
		return 0, err
	}
	f.Timestamp = inner.Timestamp
	f.Usec = inner.Usec
	f.CRealm = inner.CRealm
	f.CName = inner.CName
	f.TicketChecksum = inner.TicketChecksum
	return len(data) - len(rest), nil
}

// MarshalPAFXFastRequest encodes a KrbFastArmoredReq as the PA-FX-FAST-REQUEST
// padata-value (RFC 6113 §5.4.2):
//
//	PA-FX-FAST-REQUEST ::= CHOICE { armored-data [0] KrbFastArmoredReq, ... }
//
// A CHOICE has no SEQUENCE wrapper: the value is the chosen alternative's
// [0] EXPLICIT context element wrapping the KrbFastArmoredReq SEQUENCE.
func MarshalPAFXFastRequest(req *KrbFastArmoredReq) ([]byte, error) {
	seq, err := req.Marshal()
	if err != nil {
		return nil, err
	}
	return asn1.Marshal(asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: seq})
}

// ParsePAFXFastReply decodes a PA-FX-FAST-REPLY padata-value (RFC 6113 §5.4.3):
//
//	PA-FX-FAST-REPLY ::= CHOICE { armored-data [0] KrbFastArmoredRep, ... }
//
// It unwraps the [0] EXPLICIT alternative and parses the KrbFastArmoredRep.
func ParsePAFXFastReply(data []byte) (KrbFastArmoredRep, error) {
	var choice asn1.RawValue
	if _, err := asn1.Unmarshal(data, &choice); err != nil {
		return KrbFastArmoredRep{}, fmt.Errorf("pa-fx-fast-reply: %w", err)
	}
	if choice.Class != asn1.ClassContextSpecific || choice.Tag != 0 {
		return KrbFastArmoredRep{}, fmt.Errorf("pa-fx-fast-reply: unexpected alternative class=%d tag=%d", choice.Class, choice.Tag)
	}
	var rep KrbFastArmoredRep
	if _, err := rep.Unmarshal(choice.Bytes); err != nil {
		return KrbFastArmoredRep{}, fmt.Errorf("pa-fx-fast-reply: armored-rep: %w", err)
	}
	return rep, nil
}
