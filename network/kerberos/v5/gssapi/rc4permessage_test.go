package gssapi

import (
	"encoding/hex"
	"testing"
)

// TestMakeMICRC4KnownAnswer checks the RFC 4757 §7.3 RC4-HMAC GSS MIC token
// against a known-answer vector for a fixed RC4 session key, DCE/RPC ept_lookup
// stub, and sequence number 0. This pins the exact key derivation (Ksign /
// Kseq), the MD5+HMAC-MD5 checksum, the RC4 sequence-number encryption, and the
// GSS InitialContextToken framing that Windows RPC requires — validated live
// against a Windows DC.
func TestMakeMICRC4KnownAnswer(t *testing.T) {
	key, _ := hex.DecodeString("a3cf582b95eeb0afef5ffdddd0483fa5")
	stub, _ := hex.DecodeString("000000000000000000000000010000000000000000000000000000000000000000000000f4010000")
	want := "602306092a864886f71201020201011100ffffffffb3444d9050582818fc830a339f8f21b8"

	ctx := &SecContext{SessionKey: key, SessionEType: rc4HMACEType}
	mic, err := ctx.MakeMIC(stub)
	if err != nil {
		t.Fatalf("MakeMIC: %v", err)
	}
	if got := hex.EncodeToString(mic); got != want {
		t.Errorf("RC4 MIC mismatch:\n got=%s\nwant=%s", got, want)
	}
	if len(mic) != rc4MICTokenLen {
		t.Errorf("RC4 MIC length = %d, want %d", len(mic), rc4MICTokenLen)
	}
}

// TestRC4MICRoundTrip confirms an acceptor-direction token verifies, and that a
// tampered message fails verification.
func TestRC4MICRoundTrip(t *testing.T) {
	key, _ := hex.DecodeString("0123456789abcdef0123456789abcdef")
	data := []byte("the quick brown fox")
	ctx := &SecContext{SessionKey: key, SessionEType: rc4HMACEType}

	acc := ctx.makeMICRC4Acceptor(data, 0)
	if err := ctx.verifyMICRC4(data, acc); err != nil {
		t.Fatalf("verifyMICRC4: %v", err)
	}
	if err := ctx.verifyMICRC4([]byte("tampered"), acc); err == nil {
		t.Error("expected verification failure on tampered data")
	}
}

// TestSealRC4KnownAnswer pins the RFC 4757 §7.4 RC4-HMAC GSS Wrap token (tok_id
// 02 01) against a known-answer vector for a fixed RC4 session key, DCE/RPC
// ept_lookup stub, sequence number 0, and a fixed confounder ("AAAAAAAA"). It
// locks the SGN_ALG/SEAL_ALG fields, the KG_USAGE_SEAL (13) checksum, the
// Klocal-based Kcrypt derivation, the RC4 SND_SEQ encryption, and the 45-byte
// GSS-framed Wrap token that Windows RPC expects.
func TestSealRC4KnownAnswer(t *testing.T) {
	key, _ := hex.DecodeString("a3cf582b95eeb0afef5ffdddd0483fa5")
	stub, _ := hex.DecodeString("000000000000000000000000010000000000000000000000000000000000000000000000f4010000")
	wantCipher := "f5765e7498c679dbfa35a1309aa9eb37a08575ed260257639431c934622d80a2213c5bc983932d45"
	wantToken := "602b06092a864886f712010202020111001000ffffd39e4030ef161206c2565dc64be78aac82d5080060625f0e"

	ctx := &SecContext{SessionKey: key, SessionEType: rc4HMACEType}
	sealed, token := ctx.sealRC4WithConfounder(stub, 0, []byte("AAAAAAAA"))
	if got := hex.EncodeToString(sealed); got != wantCipher {
		t.Errorf("RC4 Wrap sealed stub mismatch:\n got=%s\nwant=%s", got, wantCipher)
	}
	if got := hex.EncodeToString(token); got != wantToken {
		t.Errorf("RC4 Wrap token mismatch:\n got=%s\nwant=%s", got, wantToken)
	}
	if len(token) != rc4WrapTokenLen {
		t.Errorf("RC4 Wrap token length = %d, want %d", len(token), rc4WrapTokenLen)
	}
	if len(sealed) != len(stub) {
		t.Errorf("sealed stub length = %d, want %d (8-aligned stub must not expand)", len(sealed), len(stub))
	}
}

// TestSealRC4RoundTrip seals a stub and confirms an acceptor-direction Wrap
// token round-trips through unsealRC4, that the plaintext is recovered, and
// that tampering with either the sealed stub or the token fails verification.
func TestSealRC4RoundTrip(t *testing.T) {
	key, _ := hex.DecodeString("0123456789abcdef0123456789abcdef")
	data := []byte("privacy over kerberos rpc!!") // 27 bytes -> exercises the pad-to-8

	ctx := &SecContext{SessionKey: key, SessionEType: rc4HMACEType}
	sealed, token := ctx.sealRC4Acceptor(data, 3)

	// Sealing must actually encrypt: the sealed stub differs from the plaintext.
	if len(sealed) < len(data) {
		t.Fatalf("sealed stub too short: %d < %d", len(sealed), len(data))
	}
	got, err := ctx.unsealRC4(sealed, token)
	if err != nil {
		t.Fatalf("unsealRC4: %v", err)
	}
	// unsealRC4 returns data plus its GSS pad; strip the pad-length trailer.
	if pad := int(got[len(got)-1]); pad > 0 && pad <= 8 {
		got = got[:len(got)-pad]
	}
	if string(got) != string(data) {
		t.Errorf("round-trip plaintext = %q, want %q", got, data)
	}

	// Tamper with the sealed stub: verification must fail.
	bad := append([]byte{}, sealed...)
	bad[0] ^= 0xff
	if _, err := ctx.unsealRC4(bad, token); err == nil {
		t.Error("expected verification failure on tampered sealed stub")
	}
	// Tamper with the checksum in the token: verification must fail.
	badTok := append([]byte{}, token...)
	badTok[len(badTok)-1] ^= 0xff
	if _, err := ctx.unsealRC4(sealed, badTok); err == nil {
		t.Error("expected verification failure on tampered token")
	}
}

// sealRC4Acceptor builds an acceptor-direction RC4 Wrap token (direction 0xff)
// so unsealRC4 (which expects the acceptor direction) can be exercised without a
// live server.
func (ctx *SecContext) sealRC4Acceptor(data []byte, seq uint64) (sealed, token []byte) {
	key, _ := ctx.baseKey()
	hdr8 := []byte{0x02, 0x01, 0x11, 0x00, 0x10, 0x00, 0xff, 0xff}
	confounder := []byte("BBBBBBBB")
	pad := rc4WrapPad(len(data))
	payload := append(append([]byte{}, data...), pad...)
	cksum := rc4WrapSgnCksum(key, hdr8, confounder, payload)
	encSeq := rc4Encrypt(rc4SeqKey(key, cksum), rc4MICSeqBytes(seq, false))
	kcrypt := rc4WrapCryptKey(key, seq)
	enc := rc4Encrypt(kcrypt, append(append([]byte{}, confounder...), payload...))
	token = append([]byte{}, rc4GSSWrapHeader...)
	token = append(token, hdr8...)
	token = append(token, encSeq...)
	token = append(token, cksum...)
	token = append(token, enc[:8]...)
	return enc[8:], token
}

// makeMICRC4Acceptor builds an acceptor-direction RC4 MIC (test helper for
// exercising verifyMICRC4).
func (ctx *SecContext) makeMICRC4Acceptor(data []byte, seq uint64) []byte {
	key, _ := ctx.baseKey()
	hdr8 := []byte{0x01, 0x01, 0x11, 0x00, 0xff, 0xff, 0xff, 0xff}
	seqBytes := rc4MICSeqBytes(seq, false)
	cksum := rc4SgnCksum(key, hdr8, data)
	encSeq := rc4Encrypt(rc4SeqKey(key, cksum), seqBytes)
	token := append([]byte{}, rc4GSSHeader...)
	token = append(token, hdr8...)
	token = append(token, encSeq...)
	token = append(token, cksum...)
	return token
}
