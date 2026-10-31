package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// referenceKDF is an independent SP800-108 counter-mode HMAC-SHA256 single-block
// computation: HMAC(key, i=1 || label || 0x00 || context || L), with label
// carrying its terminating NUL as MS-SMB2 3.1.4.2 specifies.
func referenceKDF(key []byte, label string, context []byte, bits uint32) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte{0, 0, 0, 1})
	mac.Write([]byte(label))
	mac.Write([]byte{0})
	mac.Write(context)
	mac.Write([]byte{byte(bits >> 24), byte(bits >> 16), byte(bits >> 8), byte(bits)})
	return mac.Sum(nil)[:bits/8]
}

// TestDeriveKeysAES256UsesFullSessionKey checks that for SMB 3.1.1 with an
// AES-256 cipher the encryption and decryption keys derive from the full GSS
// key (Session.FullSessionKey, MS-SMB2 3.2.5.3.1 / 3.3.5.5.3), while the signing
// and application keys keep deriving from the 16-byte SessionKey.
func TestDeriveKeysAES256UsesFullSessionKey(t *testing.T) {
	full := make([]byte, 32)
	for i := range full {
		full[i] = byte(0xA0 + i)
	}
	sessionKey := full[:16]
	preauth := bytes.Repeat([]byte{0x5C}, PreauthHashSize)

	for _, cipher := range []uint16{commands.SMB2_ENCRYPTION_AES256_CCM, commands.SMB2_ENCRYPTION_AES256_GCM} {
		k := DeriveKeysWithFullSessionKey(sessionKey, full, dialects.SMB2_DIALECT_3_1_1, preauth, cipher)
		if want := referenceKDF(full, "SMBC2SCipherKey\x00", preauth, 256); !bytes.Equal(k.ClientToServer, want) {
			t.Errorf("cipher 0x%04x: ClientToServer = %X, want %X", cipher, k.ClientToServer, want)
		}
		if want := referenceKDF(full, "SMBS2CCipherKey\x00", preauth, 256); !bytes.Equal(k.ServerToClient, want) {
			t.Errorf("cipher 0x%04x: ServerToClient = %X, want %X", cipher, k.ServerToClient, want)
		}
		if want := referenceKDF(sessionKey, "SMBSigningKey\x00", preauth, 128); !bytes.Equal(k.Signing, want) {
			t.Errorf("cipher 0x%04x: Signing = %X, want %X", cipher, k.Signing, want)
		}
		if want := referenceKDF(sessionKey, "SMBAppKey\x00", preauth, 128); !bytes.Equal(k.Application, want) {
			t.Errorf("cipher 0x%04x: Application = %X, want %X", cipher, k.Application, want)
		}
	}

	// An AES-128 cipher keeps the 16-byte SessionKey for every key.
	k := DeriveKeysWithFullSessionKey(sessionKey, full, dialects.SMB2_DIALECT_3_1_1, preauth, commands.SMB2_ENCRYPTION_AES128_GCM)
	if want := referenceKDF(sessionKey, "SMBC2SCipherKey\x00", preauth, 128); !bytes.Equal(k.ClientToServer, want) {
		t.Errorf("AES-128: ClientToServer = %X, want %X", k.ClientToServer, want)
	}
}
