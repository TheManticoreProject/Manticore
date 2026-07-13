package gssapi

import (
	"bytes"
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

// TestRC4MICUnpaddedKnownAnswer pins the RFC 4757 §7.2 RC4-HMAC GSS MIC
// SGN_CKSUM over messages whose length is NOT a multiple of four. §7.2 signs the
// data as-is: unlike the Wrap token, the MIC token does not pad the signed data.
// Each SGN_CKSUM below is the trailing 8 octets of the MIC token — the value of
// HMAC-MD5(Ksign, MD5(LE32(15) | hdr8 | data))[:8] computed per RFC 4757 §7.2
// for the fixed RC4 session key. A checksum that folded a 4-byte pad into the
// MD5 buffer would miss every one of these vectors, so this is the regression
// guard against re-introducing the pad on the MIC path.
func TestRC4MICUnpaddedKnownAnswer(t *testing.T) {
	key, _ := hex.DecodeString("1e9581d2fc64114fbbb0f7ac4502cf2a")
	ctx := &SecContext{SessionKey: key, SessionEType: rc4HMACEType}

	cases := []struct {
		data    string
		sgnCkum string // the trailing 8-octet SGN_CKSUM
	}{
		{"A", "fb9d3561dda207b2"},       // len 1, len%4 == 1
		{"BB", "eb0be530c54313a8"},      // len 2, len%4 == 2
		{"CCC", "565867a1e3d3b4d4"},     // len 3, len%4 == 3
		{"DDDDD", "fd2850c069c86363"},   // len 5, len%4 == 1
		{"EEEEEE", "5c15be411d0ef3f8"},  // len 6, len%4 == 2
		{"GGGGGGG", "12bbf34046e7e340"}, // len 7, len%4 == 3
	}

	for _, c := range cases {
		mic, err := ctx.MakeMIC([]byte(c.data))
		if err != nil {
			t.Fatalf("MakeMIC(%q): %v", c.data, err)
		}
		if len(mic) != rc4MICTokenLen {
			t.Fatalf("MIC(%q) length = %d, want %d", c.data, len(mic), rc4MICTokenLen)
		}
		got := mic[len(mic)-8:] // SGN_CKSUM is the final 8 octets
		want, _ := hex.DecodeString(c.sgnCkum)
		if !bytes.Equal(got, want) {
			t.Errorf("MIC SGN_CKSUM(%q) = %x, want %x", c.data, got, want)
		}
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

// TestRC4WrapDispatchesRC4Token confirms the generic Wrap/Unwrap API emits and
// parses the RFC 4757 §7.4 RC4-HMAC token (TOK_ID 02 01) — not the RFC 4121 CFX
// token (05 04) — for an RC4-HMAC context, and that a sealed token round-trips
// its plaintext through a peer (acceptor) context. This is the regression guard
// for the missing RC4 dispatch in Wrap/Unwrap.
func TestRC4WrapDispatchesRC4Token(t *testing.T) {
	key, _ := hex.DecodeString("0123456789abcdef0123456789abcdef")
	data := []byte("confidential payload over rc4-hmac gssapi") // 41 bytes -> pad-to-8

	init := &SecContext{SessionKey: key, SessionEType: rc4HMACEType}
	peer := &SecContext{SessionKey: key, SessionEType: rc4HMACEType, isAcceptor: true}

	tok, err := init.Wrap(data, true)
	if err != nil {
		t.Fatalf("Wrap(seal): %v", err)
	}
	// The token is GSS-framed (RFC 1964); the RC4 TOK_ID follows the OID header.
	if tok[0] != 0x60 {
		t.Fatalf("RC4 Wrap token missing GSS framing (first byte %02x)", tok[0])
	}
	tokID, rest, err := UnwrapToken(tok)
	if err != nil {
		t.Fatalf("UnwrapToken: %v", err)
	}
	if tokID != [2]byte{0x02, 0x01} {
		t.Errorf("RC4 Wrap TOK_ID = %02x %02x, want 02 01", tokID[0], tokID[1])
	}
	if tokID == [2]byte{0x05, 0x04} {
		t.Error("RC4 context must not emit a CFX (05 04) Wrap token")
	}
	// SEAL_ALG (rest[2:4], i.e. inner bytes 4..5) must be RC4 (10 00) when sealed.
	if rest[2] != 0x10 || rest[3] != 0x00 {
		t.Errorf("sealed RC4 Wrap SEAL_ALG = %02x %02x, want 10 00", rest[2], rest[3])
	}
	if bytes.Contains(tok, data) {
		t.Error("sealed RC4 Wrap token leaks plaintext")
	}

	got, sealed, err := peer.Unwrap(tok)
	if err != nil {
		t.Fatalf("Unwrap(seal): %v", err)
	}
	if !sealed {
		t.Error("Unwrap reported a sealed RC4 token as not sealed")
	}
	if !bytes.Equal(got, data) {
		t.Errorf("sealed round-trip plaintext = %q, want %q", got, data)
	}
}

// TestRC4WrapIntegrityOnlyRoundTrip confirms the integrity-only (seal=false)
// generic Wrap on an RC4-HMAC context emits an RFC 4757 token with SEAL_ALG
// "none" (ff ff) carrying the data in clear, and that a peer context recovers it.
func TestRC4WrapIntegrityOnlyRoundTrip(t *testing.T) {
	key, _ := hex.DecodeString("fedcba9876543210fedcba9876543210")
	data := []byte("integrity-only rc4 gssapi buffer!!") // exercises the pad-to-8

	init := &SecContext{SessionKey: key, SessionEType: rc4HMACEType}
	peer := &SecContext{SessionKey: key, SessionEType: rc4HMACEType, isAcceptor: true}

	tok, err := init.Wrap(data, false)
	if err != nil {
		t.Fatalf("Wrap(integrity): %v", err)
	}
	tokID, rest, err := UnwrapToken(tok)
	if err != nil {
		t.Fatalf("UnwrapToken: %v", err)
	}
	if tokID != [2]byte{0x02, 0x01} {
		t.Errorf("RC4 Wrap TOK_ID = %02x %02x, want 02 01", tokID[0], tokID[1])
	}
	// SEAL_ALG must be "none" (ff ff) and the data must travel in clear.
	if rest[2] != 0xff || rest[3] != 0xff {
		t.Errorf("integrity-only SEAL_ALG = %02x %02x, want ff ff", rest[2], rest[3])
	}
	if !bytes.Contains(tok, data) {
		t.Error("integrity-only RC4 Wrap token must carry data in clear")
	}

	got, sealed, err := peer.Unwrap(tok)
	if err != nil {
		t.Fatalf("Unwrap(integrity): %v", err)
	}
	if sealed {
		t.Error("integrity-only token reported as sealed")
	}
	if !bytes.Equal(got, data) {
		t.Errorf("integrity round-trip plaintext = %q, want %q", got, data)
	}

	// Tampering with the payload must fail the integrity check.
	bad := append([]byte{}, tok...)
	bad[len(bad)-1] ^= 0xff
	if _, _, err := peer.Unwrap(bad); err == nil {
		t.Error("expected integrity failure on tampered RC4 Wrap token")
	}
}

// TestRC4WrapAlignedPadding checks a Wrap round-trip for 8-aligned data. The
// arcfour profile has a cipher blocksize of 1, so the RFC 1964 self-describing
// pad is a single 0x01 octet regardless of length and the receiver can always
// strip it with no external length signal.
func TestRC4WrapAlignedPadding(t *testing.T) {
	key, _ := hex.DecodeString("00112233445566778899aabbccddeeff")
	data := []byte("sixteen bytes!!!") // exactly 16 bytes (8-aligned)

	init := &SecContext{SessionKey: key, SessionEType: rc4HMACEType}
	peer := &SecContext{SessionKey: key, SessionEType: rc4HMACEType, isAcceptor: true}

	for _, seal := range []bool{true, false} {
		tok, err := init.Wrap(data, seal)
		if err != nil {
			t.Fatalf("Wrap(seal=%v): %v", seal, err)
		}
		got, _, err := peer.Unwrap(tok)
		if err != nil {
			t.Fatalf("Unwrap(seal=%v): %v", seal, err)
		}
		if !bytes.Equal(got, data) {
			t.Errorf("seal=%v round-trip = %q, want %q", seal, got, data)
		}
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
