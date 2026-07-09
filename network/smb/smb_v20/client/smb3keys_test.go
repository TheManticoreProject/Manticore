package client

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
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
	deriveSMB3Keys(s, dialects.SMB2_DIALECT_3_0_0, nil)

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
	deriveSMB3Keys(s, dialects.SMB2_DIALECT_3_1_1, preauth)

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
