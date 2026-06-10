package ntlmv1

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/lm"
)

func Test_NTLMv1_FromPassword(t *testing.T) {
	serverChallenge := [8]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	tests := []struct {
		domain             string
		username           string
		password           string
		challenge          [8]byte
		expectedNTResponse string
		expectedLMResponse string
	}{
		{
			domain:             "WORKGROUP",
			username:           "Podalirius",
			password:           "Manticore",
			challenge:          serverChallenge,
			expectedNTResponse: "3675B7303C4D92F35977DBAE54CBB3198A7DF488AEECDEB9",
			expectedLMResponse: "AAB5612B2A1AC5B983EC11EBC2D33CD54BB8CF8C3E87C69A",
		},
	}

	for _, test := range tests {
		ntlmv1Ctx, err := NewNTLMv1CtxWithPassword(test.domain, test.username, test.password, test.challenge)
		if err != nil {
			t.Errorf("NewNTLMv1CtxWithPassword(%s, %s, %s, %s) returned error: %v", test.domain, test.username, test.password, test.challenge, err)
		}

		response, err := ntlmv1Ctx.ComputeResponse()
		if err != nil {
			t.Errorf("ntlmv1Ctx.ComputeResponse() returned error: %v", err)
		}
		ntResponse := response.GetNtChallengeResponse()
		hexNTResponse := hex.EncodeToString(ntResponse[:])
		if !strings.EqualFold(hexNTResponse, test.expectedNTResponse) {
			t.Errorf("ntlmv1Ctx.ComputeResponse() = %s; expected %s", hexNTResponse, test.expectedNTResponse)
		}

		lmResponse := response.GetLmChallengeResponse()
		hexLMResponse := hex.EncodeToString(lmResponse[:])
		if !strings.EqualFold(hexLMResponse, test.expectedLMResponse) {
			t.Errorf("ntlmv1Ctx.LMResponse() = %s; expected %s", hexLMResponse, test.expectedLMResponse)
		}
	}
}

// [SMB] NTLMv1-SSP Client   : 192.168.1.44
// [SMB] NTLMv1-SSP Username : THINKPAD-X61\user
// [SMB] NTLMv1-SSP Password : user
// [SMB] NTLMv1-SSP Hash     : user::THINKPAD-X61:82D6653B1699997800000000000000000000000000000000:2C84F06960FE5F2687E6D42E181F10F324A70D055A511AF7:1122334455667788

func Test_NTLMv1_ToHashcatString(t *testing.T) {
	testCases := []struct {
		name           string
		domain         string
		username       string
		password       string
		challenge      [8]byte
		expectedString string
	}{
		{
			name:           "Basic NTLMv1 Hashcat String",
			domain:         "THINKPAD-X61",
			username:       "user",
			password:       "",
			challenge:      [8]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
			expectedString: "user::THINKPAD-X61:52d536dbcefa63b9101f9c7a9d0743882f85252cc731bb25:eefabc742621a883aec4b24e0f7fbf05e17dc2880abe07cc:1122334455667788",
		},
		{
			name:           "Empty Domain NTLMv1 Hashcat String",
			domain:         "THINKPAD-X61",
			username:       "Podalirius",
			password:       "Manticore",
			challenge:      [8]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
			expectedString: "Podalirius::THINKPAD-X61:AAB5612B2A1AC5B983EC11EBC2D33CD54BB8CF8C3E87C69A:3675B7303C4D92F35977DBAE54CBB3198A7DF488AEECDEB9:1122334455667788",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ntlmv1Ctx, err := NewNTLMv1CtxWithPassword(tc.domain, tc.username, tc.password, tc.challenge)
			if err != nil {
				t.Fatalf("NewNTLMv1CtxWithPassword() error = %v", err)
			}

			response, err := ntlmv1Ctx.ComputeResponse()
			if err != nil {
				t.Fatalf("ComputeResponse() error = %v", err)
			}

			hashcatString, err := response.HashcatString()
			if err != nil {
				t.Fatalf("HashcatString() error = %v", err)
			}

			if !strings.EqualFold(hashcatString, tc.expectedString) {
				t.Errorf("HashcatString()\n\tgot  : %v\n\twant : %v", strings.ToUpper(hashcatString), strings.ToUpper(tc.expectedString))
			}
		})
	}
}

// Test_NTLMv1_LmResponse_UsesSuppliedHash verifies that a context built from a
// pre-computed LM hash (pass-the-hash) computes the LM challenge response from
// that hash, matching a context built from the password whose LM hash is equal.
func Test_NTLMv1_LmResponse_UsesSuppliedHash(t *testing.T) {
	serverChallenge := [8]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	password := "Password123!"
	lmHash := lm.LMHash(password)

	fromPassword, err := NewNTLMv1CtxWithPassword("DOMAIN", "user", password, serverChallenge)
	if err != nil {
		t.Fatalf("NewNTLMv1CtxWithPassword: %v", err)
	}
	fromHash, err := NewNTLMv1CtxWithLMHash("DOMAIN", "user", lmHash, serverChallenge)
	if err != nil {
		t.Fatalf("NewNTLMv1CtxWithLMHash: %v", err)
	}

	wantResp, err := fromPassword.ComputeLmChallengeResponse()
	if err != nil {
		t.Fatalf("ComputeLmChallengeResponse (password): %v", err)
	}
	gotResp, err := fromHash.ComputeLmChallengeResponse()
	if err != nil {
		t.Fatalf("ComputeLmChallengeResponse (hash): %v", err)
	}

	if gotResp != wantResp {
		t.Errorf("LM response from supplied hash = %x, want %x (it must not be derived from the empty password)", gotResp, wantResp)
	}
}
