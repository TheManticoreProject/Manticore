package ntlmv1

import (
	"encoding/hex"
	"strings"
	"testing"
)

// A context filled in directly with only Password set must derive both
// hashes from the password, matching NewNTLMv1CtxWithPassword.
func Test_NTLMv1_ComputeResponse_PasswordOnlyContext(t *testing.T) {
	ctx := &NTLMv1Ctx{
		Domain:          "WORKGROUP",
		Username:        "Podalirius",
		Password:        "Manticore",
		ServerChallenge: [8]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
	}

	response, err := ctx.ComputeResponse()
	if err != nil {
		t.Fatalf("ComputeResponse() returned error: %v", err)
	}

	ntResponse := response.GetNtChallengeResponse()
	if got, want := hex.EncodeToString(ntResponse[:]), "3675B7303C4D92F35977DBAE54CBB3198A7DF488AEECDEB9"; !strings.EqualFold(got, want) {
		t.Errorf("NT response = %s; expected %s", got, want)
	}
	lmResponse := response.GetLmChallengeResponse()
	if got, want := hex.EncodeToString(lmResponse[:]), "AAB5612B2A1AC5B983EC11EBC2D33CD54BB8CF8C3E87C69A"; !strings.EqualFold(got, want) {
		t.Errorf("LM response = %s; expected %s", got, want)
	}
}

func Test_NTLMv1_ComputeResponse_EmptyContextFails(t *testing.T) {
	ctx := &NTLMv1Ctx{ServerChallenge: [8]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}}
	if _, err := ctx.ComputeResponse(); err == nil {
		t.Fatal("ComputeResponse() on a context with neither NTHash nor Password returned no error")
	}
}
