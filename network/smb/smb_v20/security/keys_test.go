package security

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

// TestSMB30KeyDerivationKnownAnswer checks the SP800-108 KDF against the SMB 3.0
// key-derivation example published by Microsoft (MS-SMB2 anatomy-of-keys).
func TestSMB30KeyDerivationKnownAnswer(t *testing.T) {
	sessionKey := mustHex(t, "7CD451825D0450D235424E44BA6E78CC")

	k := DeriveKeys(sessionKey, dialects.SMB2_DIALECT_3_0_0, nil, 0)

	cases := []struct {
		name string
		got  []byte
		want string
	}{
		{"SigningKey", k.Signing, "0B7E9C5CAC36C0F6EA9AB275298CEDCE"},
		{"EncryptionKey", k.ClientToServer, "FAD27796665B313EBB578F388632B4F7"},
		{"DecryptionKey", k.ServerToClient, "B0F0427F7CEB416D1D9DCC0CD4F99447"},
		{"ApplicationKey", k.Application, "BB23A4575AA26C721AF525AF15A87B4F"},
	}
	for _, c := range cases {
		if !bytes.Equal(c.got, mustHex(t, c.want)) {
			t.Errorf("%s = %X, want %s", c.name, c.got, c.want)
		}
	}
}

// TestSMB311KeyDerivationKnownAnswer checks the SP800-108 KDF against the SMB
// 3.1.1 key-derivation example (pre-auth integrity hash as KDF context).
func TestSMB311KeyDerivationKnownAnswer(t *testing.T) {
	sessionKey := mustHex(t, "270E1BA896585EEB7AF3472D3B4C75A7")
	preauth := mustHex(t, "0DD13628CC3ED218EF9DF9772D436D0887AB9814BFAE63A80AA845F36909DB79"+
		"28622DDDAD522D9751640A459762C5A9D6BB084CBB3CE6BDADEF5D5BCE3C6C01")

	k := DeriveKeys(sessionKey, dialects.SMB2_DIALECT_3_1_1, preauth, 0)

	cases := []struct {
		name string
		got  []byte
		want string
	}{
		{"SigningKey", k.Signing, "73FE7A9A77BEF0BDE49C650D8CCB5F76"},
		{"EncryptionKey", k.ClientToServer, "629BCBC54422A0F572B97F45989B6073"},
		{"DecryptionKey", k.ServerToClient, "E2AF0DCEFAC68DA71A0DFBD0D1350D74"},
		{"ApplicationKey", k.Application, "6D7AD7954E9EC61E907B4D473DC178FF"},
	}
	for _, c := range cases {
		if !bytes.Equal(c.got, mustHex(t, c.want)) {
			t.Errorf("%s = %X, want %s", c.name, c.got, c.want)
		}
	}
}

// TestSMB311AES256KeyDerivation verifies that when an AES-256 cipher is
// negotiated the encryption and decryption keys are 32 bytes while the signing
// and application keys remain 16 bytes. It also checks that the 256-bit KDF
// output differs from the 128-bit truncation (the L value encoded into the PRF
// input changes the output even for the first 16 bytes).
func TestSMB311AES256KeyDerivation(t *testing.T) {
	sessionKey := mustHex(t, "270E1BA896585EEB7AF3472D3B4C75A7")
	preauth := mustHex(t, "0DD13628CC3ED218EF9DF9772D436D0887AB9814BFAE63A80AA845F36909DB79"+
		"28622DDDAD522D9751640A459762C5A9D6BB084CBB3CE6BDADEF5D5BCE3C6C01")

	k := DeriveKeys(sessionKey, dialects.SMB2_DIALECT_3_1_1, preauth, commands.SMB2_ENCRYPTION_AES256_GCM)

	if len(k.Signing) != 16 {
		t.Errorf("SigningKey length = %d, want 16", len(k.Signing))
	}
	if len(k.Application) != 16 {
		t.Errorf("ApplicationKey length = %d, want 16", len(k.Application))
	}
	if len(k.ClientToServer) != 32 {
		t.Errorf("EncryptionKey length = %d, want 32", len(k.ClientToServer))
	}
	if len(k.ServerToClient) != 32 {
		t.Errorf("DecryptionKey length = %d, want 32", len(k.ServerToClient))
	}

	// The signing and application keys must match the 128-bit known-answer values
	// from the test above — they are unaffected by the cipher choice.
	if want := mustHex(t, "73FE7A9A77BEF0BDE49C650D8CCB5F76"); !bytes.Equal(k.Signing, want) {
		t.Errorf("SigningKey = %X, want %X (should match AES-128 derivation)", k.Signing, want)
	}
	if want := mustHex(t, "6D7AD7954E9EC61E907B4D473DC178FF"); !bytes.Equal(k.Application, want) {
		t.Errorf("ApplicationKey = %X, want %X (should match AES-128 derivation)", k.Application, want)
	}

	// The 256-bit encryption key must differ from the 128-bit one: the L value
	// (256 vs 128) is part of the PRF input, so even the first 16 bytes change.
	enc128 := mustHex(t, "629BCBC54422A0F572B97F45989B6073")
	if bytes.Equal(k.ClientToServer[:16], enc128) {
		t.Error("EncryptionKey first 16 bytes equal the AES-128 derivation; L=256 should produce different PRF output")
	}
}

// TestSMB311GMACSigningKey verifies that GMAC uses the same 128-bit signing
// key derivation as CMAC with an AES-128 encryption cipher.
func TestSMB311GMACSigningKey(t *testing.T) {
	sessionKey := mustHex(t, "270E1BA896585EEB7AF3472D3B4C75A7")
	preauth := mustHex(t, "0DD13628CC3ED218EF9DF9772D436D0887AB9814BFAE63A80AA845F36909DB79"+
		"28622DDDAD522D9751640A459762C5A9D6BB084CBB3CE6BDADEF5D5BCE3C6C01")

	k := DeriveKeys(sessionKey, dialects.SMB2_DIALECT_3_1_1, preauth, 0)

	if len(k.Signing) != 16 {
		t.Errorf("SigningKey length = %d, want 16 for AES-GMAC", len(k.Signing))
	}
	if len(k.Application) != 16 {
		t.Errorf("ApplicationKey length = %d, want 16", len(k.Application))
	}
	if len(k.ClientToServer) != 16 {
		t.Errorf("EncryptionKey length = %d, want 16 (AES-128 cipher)", len(k.ClientToServer))
	}
	if len(k.ServerToClient) != 16 {
		t.Errorf("DecryptionKey length = %d, want 16 (AES-128 cipher)", len(k.ServerToClient))
	}

	sign128 := mustHex(t, "73FE7A9A77BEF0BDE49C650D8CCB5F76")
	if !bytes.Equal(k.Signing, sign128) {
		t.Errorf("SigningKey = %X, want %X", k.Signing, sign128)
	}
}
