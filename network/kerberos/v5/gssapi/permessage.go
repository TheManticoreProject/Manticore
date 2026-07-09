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

// ResetSendSeq overrides the running per-message send sequence. DCE/RPC uses a
// zero-based per-PDU counter for its GSS tokens rather than the AP-REQ
// authenticator sequence.
func (ctx *SecContext) ResetSendSeq(seq uint64) { ctx.sendSeq = seq }

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

// rotateRight rotates b right by n octets (RFC 4121 §4.2.5): the last n octets
// move to the front.
func rotateRight(b []byte, n int) []byte {
	if len(b) == 0 {
		return b
	}
	n %= len(b)
	out := make([]byte, len(b))
	copy(out, b[len(b)-n:])
	copy(out[n:], b[:len(b)-n])
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

// MICTokenLen returns the length of a MIC token this context emits: for RC4-HMAC
// the fixed RFC 4757 token length, otherwise the 16-byte CFX header plus the base
// key's etype checksum.
func (ctx *SecContext) MICTokenLen() int {
	_, etype := ctx.baseKey()
	if etype == rc4HMACEType {
		return rc4MICTokenLen
	}
	return 16 + micLen(etype)
}

// WrapTokenLen returns the auth_value length of a DCE-style Wrap token this
// context emits when sealing a stub of dataLen bytes. For RC4-HMAC it is the
// fixed RFC 4757 token (the sealed data is expanded in the stub, not the token,
// so dataLen is irrelevant). For an AES CFX Wrap token (RFC 4121 §4.2), the
// sealed data replaces the stub in place at the same length and the token
// carries the 16-byte header, the encrypted copy of that header, the checksum,
// the confounder and EC filler octets — so the length is
// 48 + EC + checksum, where EC pads the plaintext to the cipher block size. A
// negative dataLen requests the maximum possible length (largest EC), used to
// size the auth trailer before the exact stub length is known.
func (ctx *SecContext) WrapTokenLen(dataLen int) int {
	_, etype := ctx.baseKey()
	if etype == rc4HMACEType {
		return rc4WrapTokenLen
	}
	return 48 + wrapAESExtraCount(dataLen) + micLen(etype)
}

// wrapAESExtraCount is the EC (extra-count) filler an AES CFX Wrap token adds to
// pad the plaintext up to the 16-octet cipher block size (RFC 4121 §4.2.3). A
// negative dataLen returns the maximum EC (a full-minus-one block short), used
// when sizing the auth trailer before the stub length is fixed.
func wrapAESExtraCount(dataLen int) int {
	if dataLen < 0 {
		return 12
	}
	return (aesWrapBlock - dataLen%aesWrapBlock) & (aesWrapBlock - 1)
}

// aesWrapBlock is the cipher block size the AES CFX Wrap token pads to.
const aesWrapBlock = 16

// Seal produces a DCE-style GSS Wrap token that seals data as the context
// initiator, returning the encrypted stub (sealed in place) and the Wrap token
// for the auth_value. It is used by DCE/RPC PKT_PRIVACY; the RPC header and
// sec_trailer are not covered by the token. RC4-HMAC uses the RFC 4757 token,
// every other enctype the RFC 4121 CFX Wrap token.
func (ctx *SecContext) Seal(data []byte) (sealed, token []byte, err error) {
	if _, etype := ctx.baseKey(); etype == rc4HMACEType {
		return ctx.sealRC4(data, ctx.nextSendSeq())
	}
	return ctx.sealAES(data, ctx.nextSendSeq())
}

// Unseal decrypts and verifies a DCE-style GSS Wrap token received from the
// acceptor, returning the recovered plaintext (including any RPC auth pad the
// caller strips). RC4-HMAC uses the RFC 4757 token, every other enctype the RFC
// 4121 CFX Wrap token.
func (ctx *SecContext) Unseal(sealed, token []byte) ([]byte, error) {
	if _, etype := ctx.baseKey(); etype == rc4HMACEType {
		return ctx.unsealRC4(sealed, token)
	}
	return ctx.unsealAES(sealed, token)
}

// sealAES produces an RFC 4121 §4.2.6.2 CFX Wrap token (tok_id 05 04) with
// confidentiality, in the DCE style Windows RPC expects: the stub is encrypted
// in place at its original length while the token (auth_value) carries the
// cleartext 16-byte header followed by the rotated remainder of the ciphertext
// (the encrypted header copy, the checksum and the confounder). EC pads the
// plaintext to the cipher block size and RRC = confounder+checksum length, so
// that after the right-rotation the stub-sized run of ciphertext lands in place.
func (ctx *SecContext) sealAES(data []byte, seq uint64) (sealed, token []byte, err error) {
	key, etype := ctx.baseKey()
	ec := wrapAESExtraCount(len(data))
	rrc := aesWrapBlock + micLen(etype) // confounder + checksum length

	// The header used inside the encrypted plaintext carries EC with RRC=0
	// (RFC 4121 §4.2.4: the checksum/encryption is computed with RRC zeroed).
	hdr := wrapHeader(ctx.initiatorFlags(true), seq)
	binary.BigEndian.PutUint16(hdr[4:6], uint16(ec))

	plain := make([]byte, 0, len(data)+ec+16)
	plain = append(plain, data...)
	plain = append(plain, make([]byte, ec)...)
	plain = append(plain, hdr...)
	cipher, err := kerbcrypto.Encrypt(etype, key, kgUsageInitiatorSeal, plain)
	if err != nil {
		return nil, nil, err
	}

	// The transmitted header carries the real RRC; rotate the ciphertext right by
	// RRC+EC so the leading stub-length run can be sent in place.
	binary.BigEndian.PutUint16(hdr[6:8], uint16(rrc))
	rotated := rotateRight(cipher, rrc+ec)
	split := 16 + rrc + ec
	if split > len(rotated) {
		return nil, nil, fmt.Errorf("gssapi: AES Wrap ciphertext too short")
	}
	sealed = rotated[split:]
	token = append(append([]byte{}, hdr...), rotated[:split]...)
	return sealed, token, nil
}

// unsealAES decrypts and verifies an RFC 4121 CFX Wrap token received from the
// acceptor. sealed is the in-place encrypted stub; token is the auth_value (the
// cleartext header plus the rotated ciphertext remainder). The header's EC/RRC
// reassemble and unrotate the full ciphertext, which is decrypted with the
// acceptor-seal usage; the trailing header copy and EC filler are stripped.
func (ctx *SecContext) unsealAES(sealed, token []byte) ([]byte, error) {
	if len(token) < 16 {
		return nil, fmt.Errorf("gssapi: AES Wrap token too short")
	}
	if token[0] != tokIDWrap[0] || token[1] != tokIDWrap[1] {
		return nil, fmt.Errorf("gssapi: not a Wrap token (tok_id %02x %02x)", token[0], token[1])
	}
	if token[2]&flagSentByAcceptor == 0 {
		return nil, fmt.Errorf("gssapi: Wrap token not marked as sent by acceptor")
	}
	if token[2]&flagSealed == 0 {
		return nil, fmt.Errorf("gssapi: Wrap token is not sealed")
	}
	ec := int(binary.BigEndian.Uint16(token[4:6]))
	rrc := int(binary.BigEndian.Uint16(token[6:8]))
	key, etype := ctx.baseKey()

	// Rejoin the rotated ciphertext (token bytes after the 16-byte header, then
	// the in-place sealed stub) and undo the right-rotation.
	rotated := append(append([]byte{}, token[16:]...), sealed...)
	cipher := rotateLeft(rotated, rrc+ec)
	plain, err := kerbcrypto.Decrypt(etype, key, kgUsageAcceptorSeal, cipher)
	if err != nil {
		return nil, fmt.Errorf("gssapi: decrypt AES Wrap token: %w", err)
	}
	// plain = data | filler(ec) | header(16).
	if len(plain) < 16+ec {
		return nil, fmt.Errorf("gssapi: AES Wrap plaintext too short")
	}
	return plain[:len(plain)-16-ec], nil
}

// MakeMIC produces a MIC token over data as the context initiator. For RC4-HMAC
// it is the RFC 4757 §7.3 token; otherwise the RFC 4121 §4.2.6.1 CFX token
// (checksum(data | header) keyed with the initiator-sign usage).
func (ctx *SecContext) MakeMIC(data []byte) ([]byte, error) {
	key, etype := ctx.baseKey()
	if etype == rc4HMACEType {
		return ctx.makeMICRC4(data, ctx.nextSendSeq()), nil
	}
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
	if _, etype := ctx.baseKey(); etype == rc4HMACEType {
		return ctx.verifyMICRC4(data, token)
	}
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
	// RFC 4121 §4.2.4: the checksum is computed over data | header with EC and
	// RRC zeroed. hdr already has EC=RRC=0 here.
	sum, err := kerbcrypto.GetChecksum(ct, key, kgUsageInitiatorSeal, append(append([]byte{}, data...), hdr...))
	if err != nil {
		return nil, err
	}
	// §4.2.3: for a non-confidential Wrap token, EC encodes the checksum length.
	// Set EC and RRC to that length and right-rotate {data|checksum} by RRC so
	// the token is {header | checksum | data}, matching what AD emits/expects.
	ec := len(sum)
	binary.BigEndian.PutUint16(hdr[4:6], uint16(ec))
	binary.BigEndian.PutUint16(hdr[6:8], uint16(ec))
	body := rotateRight(append(append([]byte{}, data...), sum...), ec)
	return append(hdr, body...), nil
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
	// For a non-confidential Wrap token, EC is the checksum length (RFC 4121
	// §4.2.3); fall back to the etype's checksum length if EC is absent.
	cksumLen := ec
	if cksumLen == 0 {
		cksumLen = micLen(etype)
	}
	if len(body) < cksumLen {
		return nil, false, fmt.Errorf("gssapi: Wrap token too short for checksum")
	}
	payload := body[:len(body)-cksumLen]
	got := body[len(body)-cksumLen:]
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
