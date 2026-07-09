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
