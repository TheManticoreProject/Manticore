package gssapi

import (
	"crypto/hmac"
	"crypto/md5"
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
