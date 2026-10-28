package client

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/security"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

func TestTransformHeaderRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cipher uint16
		key    string
	}{
		{"AES-128-GCM", commands.SMB2_ENCRYPTION_AES128_GCM, "629BCBC54422A0F572B97F45989B6073"},
		{"AES-128-CCM", commands.SMB2_ENCRYPTION_AES128_CCM, "629BCBC54422A0F572B97F45989B6073"},
		{"AES-256-GCM", commands.SMB2_ENCRYPTION_AES256_GCM, "629BCBC54422A0F572B97F45989B6073A1B2C3D4E5F60718293A4B5C6D7E8F90"},
		{"AES-256-CCM", commands.SMB2_ENCRYPTION_AES256_CCM, "629BCBC54422A0F572B97F45989B6073A1B2C3D4E5F60718293A4B5C6D7E8F90"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			keyBytes := mustHex(t, tc.key)
			c := &Client{Connection: &Connection{Server: &Server{}, Dialect: dialects.SMB2_DIALECT_3_1_1, Cipher: tc.cipher}}
			c.Session = &Session{
				Client:        c,
				SessionId:     0x1234567890,
				EncryptionKey: keyBytes,
				DecryptionKey: keyBytes,
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
			if !security.IsTransformHeader(enc) {
				t.Fatalf("encrypted frame does not carry the TRANSFORM_HEADER protocol id")
			}
			if bytes.Contains(enc[security.TransformHeaderSize:], plaintext) {
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
