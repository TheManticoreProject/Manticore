package gssapi

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"encoding/binary"
	"fmt"
)

// RFC 4757 §7.3 per-message MIC token for the RC4-HMAC (arcfour) enctype. Windows
// RPC's DCE-style Kerberos uses this legacy token format (TOK_ID 01 01), wrapped
// in the GSS-API InitialContextToken framing, rather than the RFC 4121 CFX token
// (04 04). The signed data is carried out of band (the RPC stub); only the token
// travels in the auth_value.

const rc4HMACEType = 23

// rc4MICTokenLen is the fixed length of an RC4 GSS MIC token: the 13-byte GSS
// InitialContextToken header (60 23 06 09 <krb5 OID>) plus TOK_ID(2), SGN_ALG(2),
// Filler(4), SND_SEQ(8), and SGN_CKSUM(8).
const rc4MICTokenLen = 13 + 2 + 2 + 4 + 8 + 8

// rc4GSSHeader is the GSS-API InitialContextToken header for a 35-byte RC4
// per-message token: [APPLICATION 0] length 0x23, then the Kerberos 5 OID.
var rc4GSSHeader = []byte{0x60, 0x23, 0x06, 0x09, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x12, 0x01, 0x02, 0x02}

func hmacMD5(key, data []byte) []byte {
	h := hmac.New(md5.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func md5sum(data []byte) []byte {
	s := md5.Sum(data)
	return s[:]
}

func le32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

// rc4MICPad returns the RC4 MIC pad: the data is padded to a 4-byte boundary with
// the pad length byte (RFC 4757 / MIT behaviour). The pad is folded into the
// checksum only; it is not transmitted.
func rc4MICPad(n int) []byte {
	pad := (4 - (n % 4)) & 3
	out := make([]byte, pad)
	for i := range out {
		out[i] = byte(pad)
	}
	return out
}

// rc4MICSeqBytes builds the 8-byte SND_SEQ field before encryption: the 32-bit
// sequence number (big-endian) followed by the 4-byte direction indicator (0x00
// for initiator-sent tokens, 0xff for acceptor-sent tokens).
func rc4MICSeqBytes(seq uint64, initiator bool) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b[:4], uint32(seq))
	dir := byte(0x00)
	if !initiator {
		dir = 0xff
	}
	for i := 4; i < 8; i++ {
		b[i] = dir
	}
	return b
}

// recvDirByte is the SND_SEQ direction indicator the peer sets on tokens this
// context receives (RFC 4757): 0xff when the peer is the acceptor (this context
// is the initiator) and 0x00 when the peer is the initiator.
func (ctx *SecContext) recvDirByte() byte {
	if ctx.isAcceptor {
		return 0x00
	}
	return 0xff
}

// makeMICRC4 produces an RC4-HMAC GSS MIC token over data with the given sequence
// number, in the direction of this context's role (RFC 4757 §7.3).
func (ctx *SecContext) makeMICRC4(data []byte, seq uint64) []byte {
	key, _ := ctx.baseKey()
	hdr8 := []byte{0x01, 0x01, 0x11, 0x00, 0xff, 0xff, 0xff, 0xff}
	seqBytes := rc4MICSeqBytes(seq, !ctx.isAcceptor)
	cksum := rc4SgnCksum(key, hdr8, data)
	encSeq := rc4Encrypt(rc4SeqKey(key, cksum), seqBytes)

	token := make([]byte, 0, rc4MICTokenLen)
	token = append(token, rc4GSSHeader...)
	token = append(token, hdr8...)
	token = append(token, encSeq...)
	token = append(token, cksum...)
	return token
}

// verifyMICRC4 verifies an RC4-HMAC GSS MIC token received from the acceptor over
// data. It accepts the token either bare or wrapped in the InitialContextToken.
func (ctx *SecContext) verifyMICRC4(data, token []byte) error {
	// Strip the GSS InitialContextToken header (60 xx 06 09 <OID>) if present.
	if len(token) >= 13 && token[0] == 0x60 {
		token = token[13:]
	}
	if len(token) < 24 {
		return fmt.Errorf("gssapi: RC4 MIC token too short")
	}
	if token[0] != 0x01 || token[1] != 0x01 {
		return fmt.Errorf("gssapi: not an RC4 MIC token (TOK_ID %02x %02x)", token[0], token[1])
	}
	hdr8 := token[:8]
	encSeq := token[8:16]
	got := token[16:24]

	key, _ := ctx.baseKey()
	want := rc4SgnCksum(key, hdr8, data)
	if !hmac.Equal(got, want) {
		return fmt.Errorf("gssapi: RC4 MIC verification failed")
	}
	// Decrypt SND_SEQ and confirm the acceptor direction indicator.
	seq := rc4Encrypt(rc4SeqKey(key, got), encSeq)
	if seq[4] != ctx.recvDirByte() {
		return fmt.Errorf("gssapi: RC4 MIC has the wrong direction indicator")
	}
	// The token is authentic and acceptor-directed; enforce replay/sequence.
	return ctx.recvWindow.check(uint64(binary.BigEndian.Uint32(seq[:4])))
}

// rc4SgnCksum computes the RC4 MIC SGN_CKSUM: HMAC-MD5(Ksign, MD5(LE32(15) |
// tokenHeader[:8] | data | pad)), truncated to 8 bytes. Ksign is HMAC-MD5(key,
// "signaturekey\0"). The literal 15 is the MIC "sign" key-usage seed.
func rc4SgnCksum(key, hdr8, data []byte) []byte {
	ksign := hmacMD5(key, []byte("signaturekey\x00"))
	buf := make([]byte, 0, 4+8+len(data)+3)
	buf = append(buf, le32(15)...)
	buf = append(buf, hdr8...)
	buf = append(buf, data...)
	buf = append(buf, rc4MICPad(len(data))...)
	return hmacMD5(ksign, md5sum(buf))[:8]
}

// rc4SeqKey derives the RC4 key that encrypts the SND_SEQ field: HMAC-MD5(
// HMAC-MD5(key, LE32(0)), SGN_CKSUM).
func rc4SeqKey(key, sgnCksum []byte) []byte {
	return hmacMD5(hmacMD5(key, le32(0)), sgnCksum)
}

func rc4Encrypt(key, data []byte) []byte {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return nil
	}
	out := make([]byte, len(data))
	c.XORKeyStream(out, data)
	return out
}

// RFC 4757 §7.4 per-message Wrap token (TOK_ID 02 01) for the RC4-HMAC enctype.
// This is the DCE-style sealing token Windows RPC expects at
// RPC_C_AUTHN_LEVEL_PKT_PRIVACY over Kerberos. The confounder and the caller's
// data are sealed with RC4 in one keystream; the sealed data replaces the stub
// in the PDU (sealed in place), while the token header — TOK_ID, SGN_ALG,
// SEAL_ALG, Filler, SND_SEQ, SGN_CKSUM, and the encrypted confounder — travels
// in the auth_value.

// rc4WrapTokenLen is the fixed length of an RC4 GSS Wrap token as it appears in
// the auth_value: the 13-byte GSS InitialContextToken header (60 2b 06 09 <krb5
// OID>) plus TOK_ID(2), SGN_ALG(2), SEAL_ALG(2), Filler(2), SND_SEQ(8),
// SGN_CKSUM(8), and the encrypted Confounder(8). The sealed data is carried
// separately (in the stub), not in the token.
const rc4WrapTokenLen = 13 + 2 + 2 + 2 + 2 + 8 + 8 + 8

// rc4GSSWrapHeader is the GSS-API InitialContextToken header for the RC4 Wrap
// token: [APPLICATION 0] length 0x2b (the 32-byte token plus the 2+9 OID
// prefix), then the Kerberos 5 OID.
var rc4GSSWrapHeader = []byte{0x60, 0x2b, 0x06, 0x09, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x12, 0x01, 0x02, 0x02}

// rc4WrapPad pads the sealed region (confounder | data) to an 8-octet boundary
// (RFC 1964 §1.2.2.3): every pad byte holds the pad length. Although RC4 is a
// stream cipher, the arcfour GSS profile still pads to the 8-byte block used by
// the checksum. The pad is sealed and transmitted as part of the stub.
func rc4WrapPad(n int) []byte {
	pad := (8 - (n % 8)) & 7
	out := make([]byte, pad)
	for i := range out {
		out[i] = byte(pad)
	}
	return out
}

// rc4WrapSgnCksum computes the RC4 Wrap SGN_CKSUM: HMAC-MD5(Ksign, MD5(LE32(13) |
// tokenHeader[:8] | confounder | payload)), truncated to 8 bytes, where payload
// is data followed by its pad. Ksign is HMAC-MD5(key, "signaturekey\0"). The
// literal 13 is the Wrap "seal" key-usage seed (the MIC token uses 15).
func rc4WrapSgnCksum(key, hdr8, confounder, payload []byte) []byte {
	ksign := hmacMD5(key, []byte("signaturekey\x00"))
	buf := make([]byte, 0, 4+8+len(confounder)+len(payload))
	buf = append(buf, le32(13)...)
	buf = append(buf, hdr8...)
	buf = append(buf, confounder...)
	buf = append(buf, payload...)
	return hmacMD5(ksign, md5sum(buf))[:8]
}

// rc4WrapCryptKey derives the RC4 key Kcrypt that seals the confounder+data.
// Klocal is the session key with every byte XORed by 0xF0; Kcrypt is then
// HMAC-MD5(HMAC-MD5(Klocal, LE32(0)), BE32(seq)).
func rc4WrapCryptKey(key []byte, seq uint64) []byte {
	klocal := make([]byte, len(key))
	for i := range key {
		klocal[i] = key[i] ^ 0xF0
	}
	beSeq := make([]byte, 4)
	binary.BigEndian.PutUint32(beSeq, uint32(seq))
	return hmacMD5(hmacMD5(klocal, le32(0)), beSeq)
}

// sealRC4 produces an RC4-HMAC GSS Wrap token over data as the initiator (RFC
// 4757 §7.4) with the given sequence number. It returns the sealed stub (the
// encrypted data | pad, with the confounder consumed to advance the keystream)
// and the 45-byte Wrap token for the auth_value. The confounder is encrypted in
// the same RC4 keystream and carried inside the token.
func (ctx *SecContext) sealRC4(data []byte, seq uint64) (sealed, token []byte, err error) {
	confounder := make([]byte, 8)
	if _, err := rand.Read(confounder); err != nil {
		return nil, nil, err
	}
	sealed, token = ctx.sealRC4WithConfounder(data, seq, confounder)
	return sealed, token, nil
}

// sealRC4WithConfounder is the deterministic core of sealRC4 with a
// caller-supplied confounder (used both by sealRC4 with a random confounder and
// by known-answer tests with a fixed one).
func (ctx *SecContext) sealRC4WithConfounder(data []byte, seq uint64, confounder []byte) (sealed, token []byte) {
	key, _ := ctx.baseKey()
	// TOK_ID 02 01, SGN_ALG 11 00 (HMAC), SEAL_ALG 10 00 (RC4), Filler ff ff.
	hdr8 := []byte{0x02, 0x01, 0x11, 0x00, 0x10, 0x00, 0xff, 0xff}

	pad := rc4WrapPad(len(data))
	payload := make([]byte, 0, len(data)+len(pad))
	payload = append(payload, data...)
	payload = append(payload, pad...)

	cksum := rc4WrapSgnCksum(key, hdr8, confounder, payload)
	encSeq := rc4Encrypt(rc4SeqKey(key, cksum), rc4MICSeqBytes(seq, !ctx.isAcceptor))

	kcrypt := rc4WrapCryptKey(key, seq)
	plain := make([]byte, 0, 8+len(payload))
	plain = append(plain, confounder...)
	plain = append(plain, payload...)
	enc := rc4Encrypt(kcrypt, plain)

	token = make([]byte, 0, rc4WrapTokenLen)
	token = append(token, rc4GSSWrapHeader...)
	token = append(token, hdr8...)
	token = append(token, encSeq...)
	token = append(token, cksum...)
	token = append(token, enc[:8]...) // encrypted confounder
	return enc[8:], token
}

// rc4WrapPadStandalone builds the RFC 1964 §1.2.2.3 self-describing pad the
// standalone (non-DCE) RC4-HMAC GSS_Wrap token uses. The arcfour profile has a
// cipher blocksize of 1 (RC4 is a stream cipher), so "pad to the next multiple
// of the blocksize, appending between 1 and blocksize bytes" collapses to a
// single octet of value 0x01, appended unconditionally. This is what Windows /
// MIT emit and expect on the wire (verified live: an AD SASL security-layer
// offer of n data octets arrives with exactly one 0x01 pad byte). Unlike the
// DCE rc4WrapPad (which pads the sealed stub to 8 because RPC signals the true
// length out of band), a standalone token has no external length, so the pad is
// always present and self-describing for the receiver to strip.
func rc4WrapPadStandalone() []byte {
	return []byte{0x01}
}

// wrapRC4 produces a single contiguous RFC 4757 §7.4 GSS_Wrap token over data as
// this context's role, with the given sequence number. Unlike sealRC4 (which
// splits the sealed stub out for DCE in-place sealing), the whole token —
// {GSS InitialContextToken framing | TOK_ID 02 01 | SGN_ALG | SEAL_ALG | Filler |
// SND_SEQ | SGN_CKSUM | (optionally sealed) confounder | data | pad} — is
// returned contiguously, as a peer's GSS_Unwrap expects. With seal=true the
// confounder and data are RC4-encrypted (SEAL_ALG 10 00); with seal=false the
// token provides integrity only (SEAL_ALG ff ff, "none") and the confounder and
// data travel in clear.
func (ctx *SecContext) wrapRC4(data []byte, seq uint64, seal bool) ([]byte, error) {
	confounder := make([]byte, 8)
	if _, err := rand.Read(confounder); err != nil {
		return nil, err
	}
	inner := ctx.wrapRC4Inner(data, seq, confounder, seal)
	// RFC 1964 §1.2 wraps every RC4/DES per-message token in the OID-prefixed
	// GSS framing (the RFC 4121 CFX tokens dropped it). WrapToken encodes the DER
	// length, so tokens longer than the fixed 45-byte DCE form frame correctly.
	return WrapToken([2]byte{inner[0], inner[1]}, inner[2:])
}

// wrapRC4Inner builds the bare (un-framed) RFC 4757 §7.4 GSS_Wrap token with a
// caller-supplied confounder: the 32-byte header followed by the confounder and
// the padded data, sealed together when seal is set. It is the deterministic core
// of wrapRC4 (which supplies a random confounder and adds the GSS framing) and is
// driven directly by known-answer tests.
func (ctx *SecContext) wrapRC4Inner(data []byte, seq uint64, confounder []byte, seal bool) []byte {
	key, _ := ctx.baseKey()
	// TOK_ID 02 01, SGN_ALG 11 00 (HMAC), SEAL_ALG 10 00 (RC4) or ff ff (none),
	// Filler ff ff.
	hdr8 := []byte{0x02, 0x01, 0x11, 0x00, 0x10, 0x00, 0xff, 0xff}
	if !seal {
		hdr8[4], hdr8[5] = 0xff, 0xff
	}

	pad := rc4WrapPadStandalone()
	payload := make([]byte, 0, len(data)+len(pad))
	payload = append(payload, data...)
	payload = append(payload, pad...)

	cksum := rc4WrapSgnCksum(key, hdr8, confounder, payload)
	encSeq := rc4Encrypt(rc4SeqKey(key, cksum), rc4MICSeqBytes(seq, !ctx.isAcceptor))

	region := make([]byte, 0, len(confounder)+len(payload))
	region = append(region, confounder...)
	region = append(region, payload...)
	if seal {
		region = rc4Encrypt(rc4WrapCryptKey(key, seq), region)
	}

	token := make([]byte, 0, 8+8+8+len(region))
	token = append(token, hdr8...)
	token = append(token, encSeq...)
	token = append(token, cksum...)
	token = append(token, region...)
	return token
}

// unwrapRC4 parses and verifies a single contiguous RFC 4757 §7.4 GSS_Wrap token
// received from the peer, returning the recovered plaintext and whether it was
// sealed (confidential). It is the standalone counterpart of unsealRC4 (which
// takes the sealed stub and the token separately for DCE): here the confounder
// and data ride inside the one token and the self-describing pad is stripped.
func (ctx *SecContext) unwrapRC4(token []byte) (data []byte, sealed bool, err error) {
	// Strip the GSS InitialContextToken framing (RFC 1964 §1.2) if present,
	// tolerating a multi-byte DER length; UnwrapToken also verifies the mech OID.
	inner := token
	if len(token) > 0 && token[0] == 0x60 {
		tokID, rest, uerr := UnwrapToken(token)
		if uerr != nil {
			return nil, false, uerr
		}
		inner = append([]byte{tokID[0], tokID[1]}, rest...)
	}
	if len(inner) < 32 {
		return nil, false, fmt.Errorf("gssapi: RC4 Wrap token too short")
	}
	if inner[0] != 0x02 || inner[1] != 0x01 {
		return nil, false, fmt.Errorf("gssapi: not an RC4 Wrap token (TOK_ID %02x %02x)", inner[0], inner[1])
	}
	hdr8 := inner[:8]
	encSeq := inner[8:16]
	cksum := inner[16:24]
	region := inner[24:]
	// SEAL_ALG ff ff means "none" (integrity only); anything else is confidential.
	sealed = !(hdr8[4] == 0xff && hdr8[5] == 0xff)

	key, _ := ctx.baseKey()
	// Recover SND_SEQ and confirm the peer's direction indicator.
	seqBytes := rc4Encrypt(rc4SeqKey(key, cksum), encSeq)
	if seqBytes[4] != ctx.recvDirByte() {
		return nil, false, fmt.Errorf("gssapi: RC4 Wrap has the wrong direction indicator")
	}
	seq := uint64(binary.BigEndian.Uint32(seqBytes[:4]))

	if sealed {
		region = rc4Encrypt(rc4WrapCryptKey(key, seq), region)
	}
	confounder := region[:8]
	payload := region[8:]

	want := rc4WrapSgnCksum(key, hdr8, confounder, payload)
	if !hmac.Equal(cksum, want) {
		return nil, false, fmt.Errorf("gssapi: RC4 Wrap verification failed")
	}
	// Strip the self-describing RFC 1964 pad (1..8 octets, each the pad length).
	if len(payload) == 0 {
		return nil, false, fmt.Errorf("gssapi: RC4 Wrap payload missing pad")
	}
	pad := int(payload[len(payload)-1])
	if pad < 1 || pad > 8 || pad > len(payload) {
		return nil, false, fmt.Errorf("gssapi: RC4 Wrap invalid pad length %d", pad)
	}
	// The token is authentic and correctly directed; enforce replay/sequence.
	if err := ctx.recvWindow.check(seq); err != nil {
		return nil, false, err
	}
	return payload[:len(payload)-pad], sealed, nil
}

// unsealRC4 decrypts and verifies an RC4-HMAC GSS Wrap token received from the
// acceptor. sealed is the on-wire (encrypted) stub; token is the auth_value. It
// returns the recovered plaintext (data plus any GSS pad, which the caller
// strips). The sequence number and direction are recovered from the token.
func (ctx *SecContext) unsealRC4(sealed, token []byte) ([]byte, error) {
	// Strip the GSS InitialContextToken header (60 xx 06 09 <OID>) if present.
	if len(token) >= 13 && token[0] == 0x60 {
		token = token[13:]
	}
	if len(token) < 32 {
		return nil, fmt.Errorf("gssapi: RC4 Wrap token too short")
	}
	if token[0] != 0x02 || token[1] != 0x01 {
		return nil, fmt.Errorf("gssapi: not an RC4 Wrap token (TOK_ID %02x %02x)", token[0], token[1])
	}
	hdr8 := token[:8]
	encSeq := token[8:16]
	cksum := token[16:24]
	encConfounder := token[24:32]

	key, _ := ctx.baseKey()
	// Recover SND_SEQ; the acceptor direction indicator must be 0xff.
	seqBytes := rc4Encrypt(rc4SeqKey(key, cksum), encSeq)
	if seqBytes[4] != ctx.recvDirByte() {
		return nil, fmt.Errorf("gssapi: RC4 Wrap has the wrong direction indicator")
	}
	seq := uint64(binary.BigEndian.Uint32(seqBytes[:4]))

	kcrypt := rc4WrapCryptKey(key, seq)
	combined := make([]byte, 0, 8+len(sealed))
	combined = append(combined, encConfounder...)
	combined = append(combined, sealed...)
	plain := rc4Encrypt(kcrypt, combined)
	confounder := plain[:8]
	payload := plain[8:]

	want := rc4WrapSgnCksum(key, hdr8, confounder, payload)
	if !hmac.Equal(cksum, want) {
		return nil, fmt.Errorf("gssapi: RC4 Wrap verification failed")
	}
	// The token is authentic and acceptor-directed; enforce replay/sequence.
	if err := ctx.recvWindow.check(seq); err != nil {
		return nil, err
	}
	return payload, nil
}
