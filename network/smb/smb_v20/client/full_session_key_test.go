package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// TestDeriveSMB3KeysAES256UsesFullSessionKey checks that a session holding a
// 32-byte Kerberos AES256 key derives its SMB 3.1.1 AES-256 encryption and
// decryption keys from that full key (MS-SMB2 3.2.5.3.1), not from the 16-byte
// SessionKey, and its signing key from the SessionKey.
func TestDeriveSMB3KeysAES256UsesFullSessionKey(t *testing.T) {
	full := make([]byte, 32)
	for i := range full {
		full[i] = byte(i + 1)
	}
	preauth := bytes.Repeat([]byte{0x33}, 64)
	session := &Session{SessionKey: append([]byte(nil), full[:16]...), FullSessionKey: full}

	deriveSMB3Keys(session, dialects.SMB2_DIALECT_3_1_1, preauth, commands.SMB2_ENCRYPTION_AES256_GCM, 0)

	// SP800-108 counter mode with HMAC-SHA256 (MS-SMB2 3.1.4.2).
	want := func(key []byte, label string, bits uint32) []byte {
		mac := hmac.New(sha256.New, key)
		mac.Write([]byte{0, 0, 0, 1})
		mac.Write([]byte(label))
		mac.Write([]byte{0})
		mac.Write(preauth)
		mac.Write([]byte{byte(bits >> 24), byte(bits >> 16), byte(bits >> 8), byte(bits)})
		return mac.Sum(nil)[:bits/8]
	}
	if w := want(full, "SMBC2SCipherKey\x00", 256); !bytes.Equal(session.EncryptionKey, w) {
		t.Errorf("EncryptionKey = %X, want %X (derived from the full 32-byte key)", session.EncryptionKey, w)
	}
	if w := want(full, "SMBS2CCipherKey\x00", 256); !bytes.Equal(session.DecryptionKey, w) {
		t.Errorf("DecryptionKey = %X, want %X (derived from the full 32-byte key)", session.DecryptionKey, w)
	}
	if w := want(full[:16], "SMBSigningKey\x00", 128); !bytes.Equal(session.SigningKey, w) {
		t.Errorf("SigningKey = %X, want %X (derived from the 16-byte SessionKey)", session.SigningKey, w)
	}
}
