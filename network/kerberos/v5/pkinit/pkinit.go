package pkinit

import (
	"crypto/rsa"
	"crypto/sha1"
	"encoding/asn1"
	"fmt"
	"math/big"
	"time"
)

// DHNonceLen is the length in bytes of the client Diffie-Hellman nonce.
const DHNonceLen = 32

// Request bundles the client state for one PKINIT AS exchange: the ephemeral DH
// key pair and the client DH nonce. It is produced by BuildASReqPAData and
// consumed by DeriveReplyKey after the reply is parsed.
type Request struct {
	// KeyPair is the ephemeral client DH key pair.
	KeyPair *DHKeyPair
	// ClientDHNonce is the random client DH nonce sent in the AuthPack.
	ClientDHNonce []byte
}

// BuildASReqPAData builds the PA-PK-AS-REQ pre-authentication value for an AS-REQ.
//
// reqBodyDER is the DER encoding of the KDC-REQ-BODY that will be sent (the
// paChecksum is SHA-1 over exactly these bytes, RFC 4556 §3.2.1). priv/certDER
// are the client's RSA key and (self-signed, for Shadow Credentials) certificate
// used to sign the AuthPack. group selects the MODP DH group. nonce/now are the
// PKAuthenticator nonce and timestamp.
//
// It returns the PA-DATA value bytes (to be wrapped as PA-PK-AS-REQ, padata type
// 16) and a Request capturing the ephemeral DH state needed to derive the reply
// key.
func BuildASReqPAData(reqBodyDER []byte, priv *rsa.PrivateKey, certDER []byte, group DHGroup, nonce int, now time.Time) ([]byte, *Request, error) {
	kp, err := GenerateDHKeyPair(group)
	if err != nil {
		return nil, nil, err
	}
	dhNonce, err := randomBytes(DHNonceLen)
	if err != nil {
		return nil, nil, err
	}

	cpv, err := buildClientPublicValue(kp)
	if err != nil {
		return nil, nil, err
	}

	checksum := sha1.Sum(reqBodyDER)
	ap := authPack{
		PKAuthenticator: pkAuthenticator{
			CUSec:      now.Nanosecond() / 1000,
			CTime:      now.UTC().Truncate(time.Second),
			Nonce:      nonce,
			PAChecksum: checksum[:],
		},
		ClientPublicValue: cpv,
		ClientDHNonce:     dhNonce,
	}
	apDER, err := asn1.Marshal(ap)
	if err != nil {
		return nil, nil, fmt.Errorf("pkinit: marshal AuthPack: %w", err)
	}

	signed, err := BuildSignedAuthPack(apDER, priv, certDER)
	if err != nil {
		return nil, nil, err
	}

	req := paPKASReq{SignedAuthPack: signed}
	paValue, err := asn1.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("pkinit: marshal PA-PK-AS-REQ: %w", err)
	}
	return paValue, &Request{KeyPair: kp, ClientDHNonce: dhNonce}, nil
}

// Reply holds the values recovered from a PA-PK-AS-REP (dhInfo variant).
type Reply struct {
	// KDCPublicValue is the KDC's DH public value.
	KDCPublicValue *big.Int
	// ServerDHNonce is the KDC's DH nonce (may be empty).
	ServerDHNonce []byte
}

// ParseASRepPAData parses the PA-PK-AS-REP pre-authentication value (padata type
// 17), extracting the KDC's DH public value and server DH nonce from the
// dhSignedData/KDCDHKeyInfo. It only supports the Diffie-Hellman (dhInfo)
// variant.
//
// opts controls verification of the KDC's CMS SignedData signature over the
// KDCDHKeyInfo (RFC 4556 §3.2.4): when it supplies a trust anchor the signature
// and certificate chain are verified and parsing fails closed on any mismatch;
// a nil opts (or one with InsecureSkipSignatureCheck) skips that check.
func ParseASRepPAData(paValue []byte, opts *VerifyOptions) (*Reply, error) {
	// PA-PK-AS-REP ::= CHOICE { dhInfo [0] DHRepInfo, encKeyPack [1] ... }.
	var outer asn1.RawValue
	if _, err := asn1.Unmarshal(paValue, &outer); err != nil {
		return nil, fmt.Errorf("pkinit: parse PA-PK-AS-REP: %w", err)
	}
	if outer.Class != asn1.ClassContextSpecific {
		return nil, fmt.Errorf("pkinit: PA-PK-AS-REP not context-tagged (tag %d)", outer.Tag)
	}
	if outer.Tag != 0 {
		return nil, fmt.Errorf("pkinit: PA-PK-AS-REP is not the dhInfo variant (tag %d); encKeyPack is unsupported", outer.Tag)
	}

	dhSignedData, serverDHNonce, err := parseDHRepInfo(outer.Bytes)
	if err != nil {
		return nil, err
	}

	eContent, err := verifySignedDataEContent(dhSignedData, opts)
	if err != nil {
		return nil, err
	}
	var info kdcDHKeyInfo
	if _, err := asn1.Unmarshal(eContent, &info); err != nil {
		return nil, fmt.Errorf("pkinit: parse KDCDHKeyInfo: %w", err)
	}
	y, err := parseKDCPublicValue(info)
	if err != nil {
		return nil, err
	}
	return &Reply{KDCPublicValue: y, ServerDHNonce: serverDHNonce}, nil
}

// parseDHRepInfo walks the DHRepInfo content (the bytes inside the dhInfo [0]
// element), returning dhSignedData (a CMS ContentInfo DER) and the optional
// serverDHNonce. The [0] wrapper may be either an explicit wrapper around a
// DHRepInfo SEQUENCE or the SEQUENCE itself with its tag replaced; both are
// handled.
func parseDHRepInfo(dhInfoBytes []byte) (dhSignedData, serverDHNonce []byte, err error) {
	content := dhInfoBytes
	// If the content is a single SEQUENCE, the [0] tag was an explicit wrapper;
	// descend into the SEQUENCE.
	var probe asn1.RawValue
	if rest, e := asn1.Unmarshal(content, &probe); e == nil && len(rest) == 0 &&
		probe.Class == asn1.ClassUniversal && probe.Tag == asn1.TagSequence {
		content = probe.Bytes
	}

	for len(content) > 0 {
		var field asn1.RawValue
		rest, e := asn1.Unmarshal(content, &field)
		if e != nil {
			return nil, nil, fmt.Errorf("pkinit: walk DHRepInfo: %w", e)
		}
		content = rest
		if field.Class != asn1.ClassContextSpecific {
			continue
		}
		switch field.Tag {
		case 0: // dhSignedData [0] IMPLICIT OCTET STRING
			dhSignedData = field.Bytes
		case 1: // serverDHNonce [1] DHNonce (OCTET STRING), implicit or explicit
			serverDHNonce = unwrapOctetString(field)
		}
	}
	if len(dhSignedData) == 0 {
		return nil, nil, fmt.Errorf("pkinit: PA-PK-AS-REP has no dhSignedData")
	}
	return dhSignedData, serverDHNonce, nil
}

// unwrapOctetString returns the octet-string value of a context-tagged field,
// handling both the IMPLICIT form (value is field.Bytes) and the EXPLICIT form
// (field.Bytes is a wrapped universal OCTET STRING).
func unwrapOctetString(field asn1.RawValue) []byte {
	if field.IsCompound {
		var inner asn1.RawValue
		if _, err := asn1.Unmarshal(field.Bytes, &inner); err == nil &&
			inner.Class == asn1.ClassUniversal && inner.Tag == asn1.TagOctetString {
			return inner.Bytes
		}
	}
	return field.Bytes
}

// DeriveReplyKey computes the AS reply key from a parsed PKINIT reply: it derives
// the DH shared secret, then applies octetstring2key over
// (DHSharedSecret | clientDHNonce | serverDHNonce) truncated to keyLen bytes
// (RFC 4556 §3.2.3.1). keyLen is the key length of the AS-REP enctype.
func (r *Request) DeriveReplyKey(reply *Reply, keyLen int) ([]byte, error) {
	shared, err := r.KeyPair.SharedSecret(reply.KDCPublicValue)
	if err != nil {
		return nil, err
	}
	x := make([]byte, 0, len(shared)+len(r.ClientDHNonce)+len(reply.ServerDHNonce))
	x = append(x, shared...)
	x = append(x, r.ClientDHNonce...)
	x = append(x, reply.ServerDHNonce...)
	return OctetString2Key(x, keyLen), nil
}

// ReplyKeyCandidates returns the plausible octetstring2key inputs, in preference
// order, so a caller can try each until the AS-REP enc-part decrypts. RFC 4556
// says the nonces are included only when DH keys are reused and empty otherwise;
// Windows KDC behaviour varies, so the candidates cover both cases.
func (r *Request) ReplyKeyCandidates(reply *Reply, keyLen int) ([][]byte, error) {
	shared, err := r.KeyPair.SharedSecret(reply.KDCPublicValue)
	if err != nil {
		return nil, err
	}
	withNonces := append(append(append([]byte{}, shared...), r.ClientDHNonce...), reply.ServerDHNonce...)
	sharedOnly := append([]byte{}, shared...)
	return [][]byte{
		OctetString2Key(withNonces, keyLen),
		OctetString2Key(sharedOnly, keyLen),
	}, nil
}
