package message_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon/message"
)

// Golden 0x7007 frame from iSeriesNavigatorReverse
// docs/21-profile-token-generation.md ("Synthetic 69-byte vector"): single
// use, 300 s, marker DES substitute a0..a7, profile TESTUSER.
func TestGenerateTokenRequestGolden(t *testing.T) {
	want := golden(t, `00000045 0000 e009 00000000 00000000 0002 7007
		01 01
		00000007 1116 f1
		0000000a 1117 0000012c
		0000000e 1105 a0a1a2a3a4a5a6a7
		00000010 1104 e3c5e2e3e4e2c5d9 4040`)
	f, err := message.GenerateTokenRequest{
		CredentialType: message.CredentialTypeDES,
		Substitute:     []byte{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7},
		Profile:        [10]byte{0xe3, 0xc5, 0xe2, 0xe3, 0xe4, 0xe2, 0xc5, 0xd9, 0x40, 0x40},
		TokenType:      message.TokenTypeSingleUse,
		TimeoutSeconds: 300,
	}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	if got := marshal(t, f); !bytes.Equal(got, want) {
		t.Errorf("7007 =\n% x\nwant\n% x", got, want)
	}
	for name, r := range map[string]message.GenerateTokenRequest{
		"token as input":  {CredentialType: message.CredentialTypeProfileToken, Substitute: make([]byte, 32), TokenType: 0xF1, TimeoutSeconds: 1},
		"type 0xF4":       {CredentialType: message.CredentialTypeDES, Substitute: make([]byte, 8), TokenType: 0xF4, TimeoutSeconds: 1},
		"timeout 0":       {CredentialType: message.CredentialTypeDES, Substitute: make([]byte, 8), TokenType: 0xF1, TimeoutSeconds: 0},
		"timeout 3601":    {CredentialType: message.CredentialTypeDES, Substitute: make([]byte, 8), TokenType: 0xF1, TimeoutSeconds: 3601},
		"substitute size": {CredentialType: message.CredentialTypeSHA1, Substitute: make([]byte, 8), TokenType: 0xF1, TimeoutSeconds: 1},
	} {
		if _, err := r.Frame(); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestRetrieveInfoWithToken(t *testing.T) {
	token := bytes.Repeat([]byte{0x5A}, 32)
	f, err := message.RetrieveInfoRequest{CredentialType: message.CredentialTypeProfileToken, Substitute: token, Profile: [10]byte{1}}.Frame()
	if err != nil {
		t.Fatal(err)
	}
	fs, err := hostserver.ParseFields(f.Variable, hostserver.FieldPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := fs.Get(message.CPAuthToken); !ok || !bytes.Equal(v, token) || f.Template[0] != 0x02 {
		t.Errorf("token field = % x", v)
	}
	if _, ok := fs.Get(message.CPUserID); ok {
		t.Error("a token request must not carry the profile")
	}
	if _, ok := fs.Get(message.CPPasswordSubstitute); ok {
		t.Error("a token request must not carry a substitute field")
	}
	if _, err := (message.RetrieveInfoRequest{CredentialType: message.CredentialTypeProfileToken, Substitute: make([]byte, 31)}).Frame(); err == nil {
		t.Error("31-byte token accepted")
	}
}

func TestParseGenerateTokenReply(t *testing.T) {
	token := bytes.Repeat([]byte{0x77}, 32)
	ok := hostserver.Frame{Template: []byte{0, 0, 0, 0}, Variable: append([]byte{0, 0, 0, 0x26, 0x11, 0x15}, token...)}
	r, err := message.ParseGenerateTokenReply(ok)
	if err != nil || !bytes.Equal(r.Token, token) || r.Prefix != [6]byte{0, 0, 0, 0x26, 0x11, 0x15} {
		t.Fatalf("reply = %+v, %v", r, err)
	}
	tail := ok
	tail.Variable = append(append([]byte(nil), ok.Variable...), 9, 9)
	if r, err := message.ParseGenerateTokenReply(tail); err != nil || !bytes.Equal(r.Token, token) {
		t.Errorf("reply with tail = %v", err)
	}
	if r, err := message.ParseGenerateTokenReply(hostserver.Frame{Template: []byte{0, 6, 0, 2}}); err != nil || r.ReturnCode != 0x00060002 || r.Token != nil {
		t.Errorf("failure reply = %+v, %v", r, err)
	}
	short := hostserver.Frame{Template: []byte{0, 0, 0, 0}, Variable: ok.Variable[:37]}
	if _, err := message.ParseGenerateTokenReply(short); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("short reply = %v", err)
	}
}
