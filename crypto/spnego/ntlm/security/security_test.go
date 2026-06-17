package security

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/rc4"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
)

// mustHex decodes a hex string or fails the test.
func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

// TestSealKnownAnswer reproduces the GSS_WrapEx worked example from MS-NLMP 4.2.4.4.
// That example negotiates extended session security with 56-bit sealing and no key
// exchange, so the signature checksum is the raw HMAC-MD5 (not re-encrypted). Driving
// SEAL with the example's documented client signing/sealing keys and "Plaintext"
// message must produce the documented sealed data and message signature. This exercises
// the RC4 sealing stream and the MAC end to end.
func TestSealKnownAnswer(t *testing.T) {
	// Documented derived keys from MS-NLMP 4.2.4.4.
	clientSigningKey := mustHex(t, "60e799be5c72fc92922ae8ebe961fb8d")
	clientSealingKey := mustHex(t, "04dd7f014d8504d265a25cc86a3a7c06")
	// "Plaintext" as UTF-16LE.
	plaintext := mustHex(t, "50006c00610069006e007400650078007400")

	wantSealed := mustHex(t, "a02372f6530273f3aa1eb90190ce5200c99d")
	wantSig := mustHex(t, "01000000ff2aeb52f681793a00000000")

	seal, err := rc4.NewRC4WithKey(clientSealingKey)
	if err != nil {
		t.Fatalf("rc4: %v", err)
	}
	// 56-bit, no key exchange: the checksum is left unencrypted.
	ctx := &Context{
		negFlg:           flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY | flags.NTLMSSP_NEGOTIATE_56,
		clientSigningKey: clientSigningKey,
		clientSeal:       seal,
	}

	// The MS-NLMP example signs and encrypts the same buffer.
	sealed, sig := ctx.SealWith(plaintext, plaintext)
	if !bytes.Equal(sealed, wantSealed) {
		t.Errorf("sealed mismatch\n got %x\nwant %x", sealed, wantSealed)
	}
	if !bytes.Equal(sig[:], wantSig) {
		t.Errorf("signature mismatch\n got %x\nwant %x", sig[:], wantSig)
	}
}

// TestSealingKeyDerivation pins deriveKey against the documented MS-NLMP 4.2.4.4
// client sealing key: MD5(SealKey || magic) where SealKey is the 56-bit cut of the
// session key. Since deriveKey is the single function used for every signing and
// sealing key, this fixes the derivation for all of them.
func TestSealingKeyDerivation(t *testing.T) {
	sealKey56 := mustHex(t, "eb93429a8bd952") // session key cut to 56 bits
	want := mustHex(t, "04dd7f014d8504d265a25cc86a3a7c06")
	if got := deriveKey(sealKey56, clientSealingMagic); !bytes.Equal(got, want) {
		t.Errorf("client sealing key\n got %x\nwant %x", got, want)
	}
}

// TestSealingKeyInput verifies the key-strength selection of the sealing-key input.
func TestSealingKeyInput(t *testing.T) {
	key := mustHex(t, "000102030405060708090a0b0c0d0e0f")
	cases := []struct {
		name string
		flg  flags.NegotiateFlags
		want int
	}{
		{"128-bit", flags.NTLMSSP_NEGOTIATE_128, 16},
		{"56-bit", flags.NTLMSSP_NEGOTIATE_56, 7},
		{"40-bit", 0, 5},
	}
	for _, tc := range cases {
		if got := len(sealingKeyInput(key, tc.flg)); got != tc.want {
			t.Errorf("%s: input length %d, want %d", tc.name, got, tc.want)
		}
	}
}

// mirrorReceiver returns a Context whose inbound (server) keys equal the outbound
// (client) keys derived for esk/flgs, so it can verify and unseal what a sender built
// from the same parameters produces. This models the peer at the other end of the
// session, where one side's client direction is the other's server direction.
func mirrorReceiver(t *testing.T, esk []byte, flgs flags.NegotiateFlags) *Context {
	t.Helper()
	ctx, err := NewContext(esk, flgs)
	if err != nil {
		t.Fatalf("NewContext: %v", err)
	}
	seal, err := rc4.NewRC4WithKey(deriveKey(sealingKeyInput(esk, flgs), clientSealingMagic))
	if err != nil {
		t.Fatalf("rc4: %v", err)
	}
	ctx.serverSigningKey = deriveKey(esk, clientSigningMagic)
	ctx.serverSeal = seal
	return ctx
}

// TestSealUnsealRoundTrip confirms Unseal inverts Seal across several messages, with
// the sealed bytes differing from the plaintext and the sequence numbers advancing in
// lock step on both ends.
func TestSealUnsealRoundTrip(t *testing.T) {
	esk := mustHex(t, "0102030405060708090a0b0c0d0e0f10")
	flgs := flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY | flags.NTLMSSP_NEGOTIATE_128 |
		flags.NTLMSSP_NEGOTIATE_SIGN | flags.NTLMSSP_NEGOTIATE_SEAL

	sender, err := NewContext(esk, flgs)
	if err != nil {
		t.Fatalf("NewContext sender: %v", err)
	}
	receiver := mirrorReceiver(t, esk, flgs)

	messages := [][]byte{
		[]byte("first message"),
		[]byte("second, slightly longer message"),
		{},
		[]byte("fourth"),
	}
	for i, msg := range messages {
		sealed, sig := sender.SealWith(msg, msg)
		if len(msg) > 0 && bytes.Equal(sealed, msg) {
			t.Fatalf("message %d not encrypted", i)
		}
		// The receiver decrypts in place, then verifies over the recovered plaintext.
		got := append([]byte(nil), sealed...)
		receiver.DecryptInbound(got)
		if !bytes.Equal(got, msg) {
			t.Fatalf("message %d: recovered %q, want %q", i, got, msg)
		}
		if err := receiver.VerifySignature(got, sig); err != nil {
			t.Fatalf("message %d: VerifySignature: %v", i, err)
		}
	}
}

// TestSignVerifyRoundTrip confirms VerifySignature accepts a Sign output and rejects a
// tampered message, with sequence numbers advancing per message.
func TestSignVerifyRoundTrip(t *testing.T) {
	esk := mustHex(t, "1112131415161718191a1b1c1d1e1f20")
	flgs := flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY | flags.NTLMSSP_NEGOTIATE_128 |
		flags.NTLMSSP_NEGOTIATE_SIGN

	sender, err := NewContext(esk, flgs)
	if err != nil {
		t.Fatalf("NewContext sender: %v", err)
	}
	receiver := mirrorReceiver(t, esk, flgs)

	msg := []byte("the quick brown fox")
	sig := sender.Sign(msg)
	if err := receiver.VerifySignature(msg, sig); err != nil {
		t.Fatalf("VerifySignature: %v", err)
	}

	// A second message must verify under the advanced sequence number.
	msg2 := []byte("jumps over the lazy dog")
	sig2 := sender.Sign(msg2)
	if err := receiver.VerifySignature(msg2, sig2); err != nil {
		t.Fatalf("VerifySignature (2): %v", err)
	}

	// Tampering must be detected.
	bad := mirrorReceiver(t, esk, flgs)
	if err := bad.VerifySignature([]byte("not the original message"), sender.Sign(msg)); err == nil {
		t.Fatal("expected signature verification failure on tampered message")
	}
}

// TestNewContextRejectsNTLMv1 confirms the context refuses to initialize without
// extended session security, the only signing/sealing variant implemented.
func TestNewContextRejectsNTLMv1(t *testing.T) {
	key := mustHex(t, "0102030405060708090a0b0c0d0e0f10")
	if _, err := NewContext(key, flags.NTLMSSP_NEGOTIATE_128); err == nil {
		t.Fatal("expected error for non-extended-session-security flags, got nil")
	}
}
