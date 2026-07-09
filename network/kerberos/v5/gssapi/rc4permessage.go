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

// makeMICRC4 produces an RC4-HMAC GSS MIC token over data with the given sequence
// number, as the initiator (RFC 4757 §7.3).
func (ctx *SecContext) makeMICRC4(data []byte, seq uint64) []byte {
	key, _ := ctx.baseKey()
	hdr8 := []byte{0x01, 0x01, 0x11, 0x00, 0xff, 0xff, 0xff, 0xff}
	seqBytes := rc4MICSeqBytes(seq, true)
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
	if seq[4] != 0xff {
		return fmt.Errorf("gssapi: RC4 MIC not marked as sent by acceptor")
	}
	return nil
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
	encSeq := rc4Encrypt(rc4SeqKey(key, cksum), rc4MICSeqBytes(seq, true))

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
	if seqBytes[4] != 0xff {
		return nil, fmt.Errorf("gssapi: RC4 Wrap not marked as sent by acceptor")
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
	return payload, nil
}
