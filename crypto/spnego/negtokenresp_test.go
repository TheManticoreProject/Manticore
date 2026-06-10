package spnego_test

import (
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
)

func TestUnmarshalRealNegTokenResp(t *testing.T) {
	tests := []struct {
		name        string
		hexData     string
		wantState   spnego.NegState
		wantMech    []int
		wantNTLMSSP bool
		wantMIC     bool
	}{
		{
			name:        "Real SPNEGO NegTokenResp with NTLM",
			hexData:     "3081e5a0030a0101a10c060a2b06010401823702020aa281cf0481cc4e544c4d5353500002000000180018003800000005828aa2cfa9315d6d0bc75200000000000000007c007c00500000000501280a0000000f5400480049004e004b005000410044002d00580036003100020018005400480049004e004b005000410044002d00580036003100010018005400480049004e004b005000410044002d00580036003100040018007400680069006e006b007000610064002d00780036003100030018007400680069006e006b007000610064002d00780036003100060004000100000000000000",
			wantState:   spnego.NegStateAcceptIncomplete,
			wantMech:    []int{1, 3, 6, 1, 4, 1, 311, 2, 2, 10}, // NTLM OID
			wantNTLMSSP: true,
			wantMIC:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blob, err := hex.DecodeString(tt.hexData)
			if err != nil {
				t.Fatalf("Failed to decode test data: %v", err)
			}

			var resp spnego.NegTokenResp
			bytesRead, err := resp.Unmarshal(blob)
			if err != nil {
				t.Fatalf("Failed to unmarshal NegTokenResp: %v", err)
			}

			if bytesRead != len(blob) {
				t.Errorf("Expected %d bytes read, got %d", len(blob), bytesRead)
			}

			if resp.NegState != tt.wantState {
				t.Errorf("tag [0] Expected '%s', got '%s'", tt.wantState.String(), resp.NegState.String())
			}

			if tt.wantMech != nil {
				if !resp.SupportedMech.Equal(tt.wantMech) {
					t.Errorf("tag [1] Expected mech %v, got %v", tt.wantMech, resp.SupportedMech)
				}
			} else if resp.SupportedMech != nil {
				t.Error("tag [1] Expected nil SupportedMech")
			}

			if tt.wantNTLMSSP {
				if len(resp.ResponseToken) == 0 {
					t.Error("tag [2] ResponseToken is empty")
				} else if string(resp.ResponseToken[:8]) != "NTLMSSP\x00" {
					t.Errorf("tag [2] Expected NTLMSSP signature in ResponseToken, got %x", resp.ResponseToken[:8])
				}
			}

			if tt.wantMIC && resp.MechListMIC == nil {
				t.Error("tag [3] Expected MechListMIC to be present")
			} else if !tt.wantMIC && resp.MechListMIC != nil {
				t.Error("tag [3] Expected MechListMIC to be nil")
			}
		})
	}
}

func TestMarshalUnmarshalNegTokenResp(t *testing.T) {
	tests := []struct {
		name        string
		token       spnego.NegTokenResp
		shouldError bool
	}{
		{
			name: "Basic token with just NegState",
			token: spnego.NegTokenResp{
				NegState: spnego.NegStateAcceptCompleted,
			},
			shouldError: false,
		},
		{
			name: "Token with NegState and SupportedMech",
			token: spnego.NegTokenResp{
				NegState:      spnego.NegStateAcceptIncomplete,
				SupportedMech: spnego.NtlmOID,
			},
			shouldError: false,
		},
		{
			name: "Complete token with all fields",
			token: spnego.NegTokenResp{
				NegState:      spnego.NegStateAcceptCompleted,
				SupportedMech: spnego.NtlmOID,
				ResponseToken: []byte("test response token"),
				MechListMIC:   []byte("test MIC"),
			},
			shouldError: false,
		},
		{
			name: "Token with just ResponseToken",
			token: spnego.NegTokenResp{
				ResponseToken: []byte("test response token"),
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal the token
			marshaled, err := tt.token.Marshal()
			if err != nil {
				if !tt.shouldError {
					t.Fatalf("Marshal failed when it shouldn't have: %v", err)
				}
				return
			}
			if tt.shouldError {
				t.Fatalf("Marshal succeeded when it should have failed: %v", err)
			}

			// Unmarshal back
			var unmarshaled spnego.NegTokenResp
			bytesRead, err := unmarshaled.Unmarshal(marshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// Verify all bytes were read
			if bytesRead != len(marshaled) {
				t.Errorf("Not all bytes were read. Expected %d, got %d", len(marshaled), bytesRead)
			}

			// Verify fields match
			if unmarshaled.NegState != tt.token.NegState {
				t.Errorf("NegState mismatch. Expected %v, got %v", tt.token.NegState, unmarshaled.NegState)
			}

			if !unmarshaled.SupportedMech.Equal(tt.token.SupportedMech) {
				t.Errorf("SupportedMech mismatch. Expected %v, got %v", tt.token.SupportedMech, unmarshaled.SupportedMech)
			}

			if string(unmarshaled.ResponseToken) != string(tt.token.ResponseToken) {
				t.Errorf("ResponseToken mismatch. Expected %v, got %v", tt.token.ResponseToken, unmarshaled.ResponseToken)
			}

			if string(unmarshaled.MechListMIC) != string(tt.token.MechListMIC) {
				t.Errorf("MechListMIC mismatch. Expected %v, got %v", tt.token.MechListMIC, unmarshaled.MechListMIC)
			}
		})
	}
}

// TestUnmarshalRejectsEmptyNegStateEnum verifies that a NegTokenResp whose
// negState ENUMERATED carries no content bytes is rejected with an error rather
// than panicking on an out-of-range index.
func TestUnmarshalRejectsEmptyNegStateEnum(t *testing.T) {
	// SEQUENCE { [0] EXPLICIT { ENUMERATED (length 0) } }
	//   30 04        SEQUENCE, length 4
	//     a0 02      [0] context, length 2
	//       0a 00    ENUMERATED, length 0 (no content)
	malformed := []byte{0x30, 0x04, 0xA0, 0x02, 0x0A, 0x00}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NegTokenResp.Unmarshal panicked on an empty negState ENUMERATED: %v", r)
		}
	}()

	resp := &spnego.NegTokenResp{}
	if _, err := resp.Unmarshal(malformed); err == nil {
		t.Fatal("Unmarshal should reject an empty negState ENUMERATED, got nil error")
	}
}
