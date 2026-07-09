package client

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// TestSMB3CMACSignAndVerify checks the AES-128-CMAC signing round-trip used by
// the SMB 3.x dialects: a signed message verifies with the right key, sets the
// SIGNED flag, and fails to verify under a wrong key or after tampering.
func TestSMB3CMACSignAndVerify(t *testing.T) {
	key := mustHex(t, "0B7E9C5CAC36C0F6EA9AB275298CEDCE")

	m := message.NewMessage()
	m.Header.MessageId = 7
	m.SetCommand(commands.NewTreeDisconnectRequest())
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	signMessageForDialect(dialects.SMB2_DIALECT_3_1_1, key, wire)

	decoded := message.NewMessage()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.Header.Flags.IsSigned() {
		t.Errorf("SMB2_FLAGS_SIGNED not set after CMAC signing")
	}

	if !verifySignatureForDialect(dialects.SMB2_DIALECT_3_1_1, key, wire) {
		t.Errorf("CMAC signature failed to verify with the correct key")
	}
	if verifySignatureForDialect(dialects.SMB2_DIALECT_3_1_1, mustHex(t, "00000000000000000000000000000000"), wire) {
		t.Errorf("CMAC signature verified with a wrong key")
	}
	wire[len(wire)-1] ^= 0xFF
	if verifySignatureForDialect(dialects.SMB2_DIALECT_3_1_1, key, wire) {
		t.Errorf("CMAC signature verified for a tampered message")
	}
}

// TestCCMKnownAnswerRFC3610 checks the AES-CCM implementation against the first
// test vector from RFC 3610 (M=8, L=2, 8-byte AAD).
func TestCCMKnownAnswerRFC3610(t *testing.T) {
	key := mustHex(t, "C0C1C2C3C4C5C6C7C8C9CACBCCCDCECF")
	nonce := mustHex(t, "00000003020100A0A1A2A3A4A5") // 13 bytes -> L=2
	aad := mustHex(t, "0001020304050607")
	plaintext := mustHex(t, "08090A0B0C0D0E0F101112131415161718191A1B1C1D1E")
	// Expected ciphertext || 8-byte MAC from RFC 3610 packet vector #1.
	wantCT := mustHex(t, "588C979A61C663D2F066D0C2C0F989806D5F6B61DAC384")
	wantMAC := mustHex(t, "17E8D12CFDF926E0")

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes: %v", err)
	}

	// The RFC 3610 vector uses an 8-byte tag; exercise the generic core through a
	// dedicated tag length rather than the SMB-fixed 16.
	tag := ccmTag(block, nonce, plaintext, aad, 15-len(nonce), len(wantMAC))
	var counter0 [16]byte
	ccmCounterBlock(counter0[:], nonce, 15-len(nonce), 0)
	var s0 [16]byte
	block.Encrypt(s0[:], counter0[:])
	gotMAC := make([]byte, 8)
	for i := range gotMAC {
		gotMAC[i] = tag[i] ^ s0[i]
	}
	if !bytes.Equal(gotMAC, wantMAC) {
		t.Errorf("CCM MAC = %X, want %X", gotMAC, wantMAC)
	}

	ct := make([]byte, len(plaintext))
	ccmCTR(block, nonce, 15-len(nonce), plaintext, ct)
	if !bytes.Equal(ct, wantCT) {
		t.Errorf("CCM ciphertext = %X, want %X", ct, wantCT)
	}
}

// TestCCMSealOpenRoundTrip checks the SMB-shaped AES-128-CCM AEAD (11-byte
// nonce, 16-byte tag) seals and opens, and rejects tampering.
func TestCCMSealOpenRoundTrip(t *testing.T) {
	key := mustHex(t, "00112233445566778899AABBCCDDEEFF")
	block, _ := aes.NewCipher(key)
	nonce := mustHex(t, "0102030405060708090A0B") // 11 bytes
	aad := []byte("transform-header-aad-32bytes!!!!")
	plaintext := []byte("the quick brown fox jumps over the lazy SMB dog")

	sealed, err := ccmSeal(block, nonce, plaintext, aad)
	if err != nil {
		t.Fatalf("ccmSeal: %v", err)
	}
	if len(sealed) != len(plaintext)+ccmTagSize {
		t.Fatalf("sealed length = %d, want %d", len(sealed), len(plaintext)+ccmTagSize)
	}
	got, err := ccmOpen(block, nonce, sealed, aad)
	if err != nil {
		t.Fatalf("ccmOpen: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("ccmOpen = %q, want %q", got, plaintext)
	}
	sealed[0] ^= 0xFF
	if _, err := ccmOpen(block, nonce, sealed, aad); err == nil {
		t.Errorf("ccmOpen accepted a tampered ciphertext")
	}
}

// TestTransformHeaderRoundTrip checks that a message encrypted into an SMB2
// TRANSFORM_HEADER by the client decrypts back to the original, for both the
// AES-128-GCM and AES-128-CCM ciphers.
func TestTransformHeaderRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cipher uint16
	}{
		{"AES-128-GCM", commands.SMB2_ENCRYPTION_AES128_GCM},
		{"AES-128-CCM", commands.SMB2_ENCRYPTION_AES128_CCM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &Client{Connection: &Connection{Server: &Server{}, Dialect: dialects.SMB2_DIALECT_3_1_1, Cipher: tc.cipher}}
			c.Session = &Session{
				Client:        c,
				SessionId:     0x1234567890,
				EncryptionKey: mustHex(t, "629BCBC54422A0F572B97F45989B6073"),
				DecryptionKey: mustHex(t, "629BCBC54422A0F572B97F45989B6073"),
				EncryptData:   true,
			}

			m := message.NewMessage()
			m.Header.MessageId = 5
			m.Header.SessionId = c.Session.SessionId
			m.SetCommand(commands.NewTreeDisconnectRequest())
			plaintext, err := m.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			enc, err := c.encryptMessage(plaintext)
			if err != nil {
				t.Fatalf("encryptMessage: %v", err)
			}
			if !isTransformHeader(enc) {
				t.Fatalf("encrypted frame does not carry the TRANSFORM_HEADER protocol id")
			}
			if bytes.Contains(enc[transformHeaderSize:], plaintext) {
				t.Errorf("plaintext leaked into the encrypted body")
			}

			dec, err := c.decryptMessage(enc)
			if err != nil {
				t.Fatalf("decryptMessage: %v", err)
			}
			if !bytes.Equal(dec, plaintext) {
				t.Errorf("round-trip mismatch:\n got %X\nwant %X", dec, plaintext)
			}

			// Tampering with the ciphertext must be detected by the AEAD tag.
			enc[len(enc)-1] ^= 0xFF
			if _, err := c.decryptMessage(enc); err == nil {
				t.Errorf("decryptMessage accepted a tampered frame")
			}
		})
	}
}

// TestNegotiateContextRoundTrip checks that the SMB 3.1.1 negotiate contexts
// marshal and parse back with 8-byte alignment preserved, and that the
// selected-cipher / selected-hash helpers read them.
func TestNegotiateContextRoundTrip(t *testing.T) {
	req := commands.NewNegotiateRequest()
	for _, d := range []dialects.Dialect{
		dialects.SMB2_DIALECT_2_0_2, dialects.SMB2_DIALECT_2_1_0,
		dialects.SMB2_DIALECT_3_0_0, dialects.SMB2_DIALECT_3_0_2, dialects.SMB2_DIALECT_3_1_1,
	} {
		req.AddDialect(d)
	}
	salt := mustHex(t, "FA49E6578F1F3A9F4CD3E9CC14A67AA884B3D05844E0E5A118225C15887F32FF")
	req.Contexts = []*commands.NegotiateContext{
		commands.NewPreauthIntegrityContext(salt),
		commands.NewEncryptionContext([]uint16{commands.SMB2_ENCRYPTION_AES128_GCM, commands.SMB2_ENCRYPTION_AES128_CCM}),
	}

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	parsed := commands.NewNegotiateRequest()
	if _, err := parsed.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(parsed.Contexts) != 2 {
		t.Fatalf("parsed %d contexts, want 2", len(parsed.Contexts))
	}
	if got := commands.SelectedPreauthHash(parsed.Contexts); got != commands.SMB2_PREAUTH_HASH_SHA_512 {
		t.Errorf("preauth hash = 0x%04x, want SHA-512", got)
	}
	// The request's encryption context lists GCM first.
	if got := commands.SelectedCipher(parsed.Contexts); got != commands.SMB2_ENCRYPTION_AES128_GCM {
		t.Errorf("selected cipher = 0x%04x, want AES-128-GCM", got)
	}
}

// TestPreauthHashMatchesPublishedVector reproduces the published SMB 3.1.1
// NEGOTIATE-request pre-auth hash step: SHA-512(zeros(64) || request) must equal
// the documented value.
func TestPreauthHashMatchesPublishedVector(t *testing.T) {
	negReq := mustHex(t, ""+
		"FE534D4240000100000000000000800000000000000000000100000000000000FFFE000000000000"+
		"00000000000000000000000000000000000000000000000024000500000000003F000000ECD86F32"+
		"6276024F9F7752B89BB33F3A70000000020000000202100200030203110300000100260000000000"+
		"010020000100FA49E6578F1F3A9F4CD3E9CC14A67AA884B3D05844E0E5A118225C15887F32FF0000"+
		"0200060000000000020002000100")
	want := mustHex(t, ""+
		"DD94EFC5321BB618A2E208BA8920D2F422992526947A409B5037DE1E0FE8C736"+
		"2B8C47122594CDE0CE26AA9DFC8BCDBDE0621957672623351A7540F1E54A0426")

	got := preauthUpdate(make([]byte, preauthHashLength), negReq)
	if !bytes.Equal(got, want) {
		t.Errorf("pre-auth hash mismatch:\n got %s\nwant %s", hex.EncodeToString(got), hex.EncodeToString(want))
	}
}
