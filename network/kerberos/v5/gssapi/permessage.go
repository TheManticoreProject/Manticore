package gssapi

import (
	"crypto/hmac"
	"encoding/binary"
	"fmt"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
)

// RFC 4121 §4.2.6 per-message tokens (MIC and Wrap), for the RFC 3961/3962/8009
// (AES) crypto profiles. These provide message integrity (GetMIC/VerifyMIC) and
// confidentiality+integrity (Wrap/Unwrap) once a context is established, and are
// what SMB and LDAP use for signing and sealing.

// RFC 4121 §2 key-usage numbers for per-message tokens.
const (
	kgUsageAcceptorSeal  = 22
	kgUsageAcceptorSign  = 23
	kgUsageInitiatorSeal = 24
	kgUsageInitiatorSign = 25
)

// Per-message token Flags bits (RFC 4121 §4.2.2).
const (
	flagSentByAcceptor = 0x01
	flagSealed         = 0x02
	flagAcceptorSubkey = 0x04
)

var (
	tokIDMIC  = [2]byte{0x04, 0x04}
	tokIDWrap = [2]byte{0x05, 0x04}
)

// baseKey returns the key protecting per-message tokens: the acceptor subkey if
// one was asserted, otherwise the ticket session key.
func (ctx *SecContext) baseKey() ([]byte, int) {
	if len(ctx.SubKey) > 0 {
		return ctx.SubKey, ctx.SubKeyEType
	}
	return ctx.SessionKey, ctx.SessionEType
}

// initiatorFlags is the Flags byte for tokens this (initiator) context sends.
func (ctx *SecContext) initiatorFlags(sealed bool) byte {
	var f byte
	if ctx.acceptorSubkey {
		f |= flagAcceptorSubkey
	}
	if sealed {
		f |= flagSealed
	}
	return f // SentByAcceptor stays 0 (we are the initiator)
}

func (ctx *SecContext) nextSendSeq() uint64 {
	s := ctx.sendSeq
	ctx.sendSeq++
	return s
}

// micLen returns the checksum length appended to MIC / integrity-only Wrap
// tokens for an etype.
func micLen(etype int) int {
	if ct, ok := kerbcrypto.ChecksumTypeForEType(etype); ok {
		switch ct {
		case 15, 16: // hmac-sha1-96-aes*
			return 12
		case 19: // hmac-sha256-128
			return 16
		case 20: // hmac-sha384-192
			return 24
		default: // hmac-md5 (rc4)
			return 16
		}
	}
	return 0
}

// rotateLeft rotates b left by n octets (undoing a right rotation by n).
func rotateLeft(b []byte, n int) []byte {
	if len(b) == 0 {
		return b
	}
	n %= len(b)
	out := make([]byte, len(b))
	copy(out, b[n:])
	copy(out[len(b)-n:], b[:n])
	return out
}

// micHeader builds the 16-byte MIC token header.
func micHeader(flags byte, seq uint64) []byte {
	h := make([]byte, 16)
	h[0], h[1] = tokIDMIC[0], tokIDMIC[1]
	h[2] = flags
	for i := 3; i < 8; i++ {
		h[i] = 0xFF
	}
	binary.BigEndian.PutUint64(h[8:], seq)
	return h
}

// wrapHeader builds the 16-byte Wrap token header with EC=0, RRC=0.
func wrapHeader(flags byte, seq uint64) []byte {
	h := make([]byte, 16)
	h[0], h[1] = tokIDWrap[0], tokIDWrap[1]
	h[2] = flags
	h[3] = 0xFF
	// EC (4..5) and RRC (6..7) left as 0.
	binary.BigEndian.PutUint64(h[8:], seq)
	return h
}

// MakeMIC produces a MIC token over data as the context initiator (RFC 4121
// §4.2.6.1): checksum(data | header) keyed with the initiator-sign usage.
func (ctx *SecContext) MakeMIC(data []byte) ([]byte, error) {
	key, etype := ctx.baseKey()
	ct, ok := kerbcrypto.ChecksumTypeForEType(etype)
	if !ok {
		return nil, fmt.Errorf("gssapi: no checksum for etype %d", etype)
	}
	hdr := micHeader(ctx.initiatorFlags(false), ctx.nextSendSeq())
	sum, err := kerbcrypto.GetChecksum(ct, key, kgUsageInitiatorSign, append(append([]byte{}, data...), hdr...))
	if err != nil {
		return nil, err
	}
	return append(hdr, sum...), nil
}

// VerifyMIC verifies a MIC token received from the acceptor over data.
func (ctx *SecContext) VerifyMIC(data, token []byte) error {
	if len(token) < 16 {
		return fmt.Errorf("gssapi: MIC token too short")
	}
	if token[0] != tokIDMIC[0] || token[1] != tokIDMIC[1] {
		return fmt.Errorf("gssapi: not a MIC token")
	}
	if token[2]&flagSentByAcceptor == 0 {
		return fmt.Errorf("gssapi: MIC token not marked as sent by acceptor")
	}
	key, etype := ctx.baseKey()
	ct, ok := kerbcrypto.ChecksumTypeForEType(etype)
	if !ok {
		return fmt.Errorf("gssapi: no checksum for etype %d", etype)
	}
	hdr := token[:16]
	got := token[16:]
	want, err := kerbcrypto.GetChecksum(ct, key, kgUsageAcceptorSign, append(append([]byte{}, data...), hdr...))
	if err != nil {
		return err
	}
	if !hmac.Equal(got, want) {
		return fmt.Errorf("gssapi: MIC verification failed")
	}
	return nil
}

// Wrap produces a Wrap token over data as the context initiator (RFC 4121
// §4.2.6.2). With seal=true the token provides confidentiality+integrity;
// otherwise integrity only. EC and RRC are 0 (no filler, no rotation).
func (ctx *SecContext) Wrap(data []byte, seal bool) ([]byte, error) {
	key, etype := ctx.baseKey()
	hdr := wrapHeader(ctx.initiatorFlags(seal), ctx.nextSendSeq())

	if seal {
		pt := append(append([]byte{}, data...), hdr...)
		ct, err := kerbcrypto.Encrypt(etype, key, kgUsageInitiatorSeal, pt)
		if err != nil {
			return nil, err
		}
		return append(hdr, ct...), nil
	}

	ct, ok := kerbcrypto.ChecksumTypeForEType(etype)
	if !ok {
		return nil, fmt.Errorf("gssapi: no checksum for etype %d", etype)
	}
	sum, err := kerbcrypto.GetChecksum(ct, key, kgUsageInitiatorSeal, append(append([]byte{}, data...), hdr...))
	if err != nil {
		return nil, err
	}
	out := append([]byte{}, hdr...)
	out = append(out, data...)
	out = append(out, sum...)
	return out, nil
}

// Unwrap decodes a Wrap token received from the acceptor, returning the
// plaintext data and whether it was sealed (confidential).
func (ctx *SecContext) Unwrap(token []byte) (data []byte, sealed bool, err error) {
	if len(token) < 16 {
		return nil, false, fmt.Errorf("gssapi: Wrap token too short")
	}
	if token[0] != tokIDWrap[0] || token[1] != tokIDWrap[1] {
		return nil, false, fmt.Errorf("gssapi: not a Wrap token")
	}
	flags := token[2]
	if flags&flagSentByAcceptor == 0 {
		return nil, false, fmt.Errorf("gssapi: Wrap token not marked as sent by acceptor")
	}
	sealed = flags&flagSealed != 0
	ec := int(binary.BigEndian.Uint16(token[4:6]))
	rrc := int(binary.BigEndian.Uint16(token[6:8]))
	hdr := token[:16]
	body := rotateLeft(token[16:], rrc)

	key, etype := ctx.baseKey()

	if sealed {
		pt, err := kerbcrypto.Decrypt(etype, key, kgUsageAcceptorSeal, body)
		if err != nil {
			return nil, false, fmt.Errorf("gssapi: decrypt Wrap token: %w", err)
		}
		// pt = data | filler(ec) | header(16).
		if len(pt) < 16+ec {
			return nil, false, fmt.Errorf("gssapi: Wrap plaintext too short")
		}
		return pt[:len(pt)-16-ec], true, nil
	}

	ct, ok := kerbcrypto.ChecksumTypeForEType(etype)
	if !ok {
		return nil, false, fmt.Errorf("gssapi: no checksum for etype %d", etype)
	}
	ml := micLen(etype)
	if len(body) < ml {
		return nil, false, fmt.Errorf("gssapi: Wrap token too short for checksum")
	}
	payload := body[:len(body)-ml]
	got := body[len(body)-ml:]
	// The checksum is over data | header with EC and RRC zeroed.
	zhdr := append([]byte{}, hdr...)
	zhdr[4], zhdr[5], zhdr[6], zhdr[7] = 0, 0, 0, 0
	want, err := kerbcrypto.GetChecksum(ct, key, kgUsageAcceptorSeal, append(append([]byte{}, payload...), zhdr...))
	if err != nil {
		return nil, false, err
	}
	if !hmac.Equal(got, want) {
		return nil, false, fmt.Errorf("gssapi: Wrap integrity check failed")
	}
	return payload, false, nil
}
