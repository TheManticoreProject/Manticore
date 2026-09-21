package client

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

	s := &Session{SessionKey: sessionKey}
	deriveSMB3Keys(s, dialects.SMB2_DIALECT_3_0_0, nil, 0, -1)

	cases := []struct {
		name string
		got  []byte
		want string
	}{
		{"SigningKey", s.SigningKey, "0B7E9C5CAC36C0F6EA9AB275298CEDCE"},
		{"EncryptionKey", s.EncryptionKey, "FAD27796665B313EBB578F388632B4F7"},
		{"DecryptionKey", s.DecryptionKey, "B0F0427F7CEB416D1D9DCC0CD4F99447"},
		{"ApplicationKey", s.ApplicationKey, "BB23A4575AA26C721AF525AF15A87B4F"},
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

	s := &Session{SessionKey: sessionKey}
	deriveSMB3Keys(s, dialects.SMB2_DIALECT_3_1_1, preauth, 0, -1)

	cases := []struct {
		name string
		got  []byte
		want string
	}{
		{"SigningKey", s.SigningKey, "73FE7A9A77BEF0BDE49C650D8CCB5F76"},
		{"EncryptionKey", s.EncryptionKey, "629BCBC54422A0F572B97F45989B6073"},
		{"DecryptionKey", s.DecryptionKey, "E2AF0DCEFAC68DA71A0DFBD0D1350D74"},
		{"ApplicationKey", s.ApplicationKey, "6D7AD7954E9EC61E907B4D473DC178FF"},
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

	s := &Session{SessionKey: sessionKey}
	deriveSMB3Keys(s, dialects.SMB2_DIALECT_3_1_1, preauth, commands.SMB2_ENCRYPTION_AES256_GCM, -1)

	if len(s.SigningKey) != 16 {
		t.Errorf("SigningKey length = %d, want 16", len(s.SigningKey))
	}
	if len(s.ApplicationKey) != 16 {
		t.Errorf("ApplicationKey length = %d, want 16", len(s.ApplicationKey))
	}
	if len(s.EncryptionKey) != 32 {
		t.Errorf("EncryptionKey length = %d, want 32", len(s.EncryptionKey))
	}
	if len(s.DecryptionKey) != 32 {
		t.Errorf("DecryptionKey length = %d, want 32", len(s.DecryptionKey))
	}

	// The signing and application keys must match the 128-bit known-answer values
	// from the test above — they are unaffected by the cipher choice.
	if want := mustHex(t, "73FE7A9A77BEF0BDE49C650D8CCB5F76"); !bytes.Equal(s.SigningKey, want) {
		t.Errorf("SigningKey = %X, want %X (should match AES-128 derivation)", s.SigningKey, want)
	}
	if want := mustHex(t, "6D7AD7954E9EC61E907B4D473DC178FF"); !bytes.Equal(s.ApplicationKey, want) {
		t.Errorf("ApplicationKey = %X, want %X (should match AES-128 derivation)", s.ApplicationKey, want)
	}

	// The 256-bit encryption key must differ from the 128-bit one: the L value
	// (256 vs 128) is part of the PRF input, so even the first 16 bytes change.
	enc128 := mustHex(t, "629BCBC54422A0F572B97F45989B6073")
	if bytes.Equal(s.EncryptionKey[:16], enc128) {
		t.Error("EncryptionKey first 16 bytes equal the AES-128 derivation; L=256 should produce different PRF output")
	}
}

// TestSMB311GMACSigning256BitKey verifies that when AES-GMAC is the negotiated
// signing algorithm the signing key is 32 bytes while all other keys retain
// their normal lengths. The 256-bit signing key must also differ from the
// 128-bit one (the L value changes the PRF output).
func TestSMB311GMACSigning256BitKey(t *testing.T) {
	sessionKey := mustHex(t, "270E1BA896585EEB7AF3472D3B4C75A7")
	preauth := mustHex(t, "0DD13628CC3ED218EF9DF9772D436D0887AB9814BFAE63A80AA845F36909DB79"+
		"28622DDDAD522D9751640A459762C5A9D6BB084CBB3CE6BDADEF5D5BCE3C6C01")

	s := &Session{SessionKey: sessionKey}
	deriveSMB3Keys(s, dialects.SMB2_DIALECT_3_1_1, preauth, 0, commands.SMB2_SIGNING_ALG_AES_GMAC)

	if len(s.SigningKey) != 32 {
		t.Errorf("SigningKey length = %d, want 32 for AES-GMAC", len(s.SigningKey))
	}
	if len(s.ApplicationKey) != 16 {
		t.Errorf("ApplicationKey length = %d, want 16", len(s.ApplicationKey))
	}
	if len(s.EncryptionKey) != 16 {
		t.Errorf("EncryptionKey length = %d, want 16 (AES-128 cipher)", len(s.EncryptionKey))
	}
	if len(s.DecryptionKey) != 16 {
		t.Errorf("DecryptionKey length = %d, want 16 (AES-128 cipher)", len(s.DecryptionKey))
	}

	sign128 := mustHex(t, "73FE7A9A77BEF0BDE49C650D8CCB5F76")
	if bytes.Equal(s.SigningKey[:16], sign128) {
		t.Error("SigningKey first 16 bytes equal the AES-CMAC derivation; L=256 should produce different PRF output")
	}
}
