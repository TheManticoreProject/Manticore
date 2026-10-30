package client

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func TestSetAuthValidation(t *testing.T) {
	creds, err := credentials.NewCredentials("", "user", "pass", "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}

	cases := []struct {
		name     string
		authType uint8
		level    uint8
		creds    *credentials.Credentials
		wantErr  bool
	}{
		{"ntlm privacy", pdu.AuthTypeNTLMSSP, pdu.AuthLevelPktPrivacy, creds, false},
		{"ntlm integrity", pdu.AuthTypeNTLMSSP, pdu.AuthLevelPktIntegrity, creds, false},
		{"ntlm pkt", pdu.AuthTypeNTLMSSP, pdu.AuthLevelPkt, creds, false},
		{"ntlm call (promoted to pkt)", pdu.AuthTypeNTLMSSP, pdu.AuthLevelCall, creds, false},
		{"ntlm connect", pdu.AuthTypeNTLMSSP, pdu.AuthLevelConnect, creds, false},
		{"unsupported auth type", pdu.AuthTypeGSSKerberos, pdu.AuthLevelPktPrivacy, creds, true},
		{"unsupported level (none)", pdu.AuthTypeNTLMSSP, pdu.AuthLevelNone, creds, true},
		{"unsupported level (default)", pdu.AuthTypeNTLMSSP, pdu.AuthLevelDefault, creds, true},
		{"nil credentials", pdu.AuthTypeNTLMSSP, pdu.AuthLevelPktPrivacy, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(nil)
			err := c.SetAuth(tc.authType, tc.level, tc.creds)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if c.authConfigured() {
					t.Fatal("auth must not be configured after a rejected SetAuth")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !c.authConfigured() {
				t.Fatal("auth should be configured after a successful SetAuth")
			}
		})
	}
}

// TestSetAuthCallPromotedToPkt verifies that the connection-oriented client maps
// RPC_C_AUTHN_LEVEL_CALL onto PKT, which is what is carried in the sec_trailer.
func TestSetAuthCallPromotedToPkt(t *testing.T) {
	creds, err := credentials.NewCredentials("", "user", "pass", "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}
	c := NewClient(nil)
	if err := c.SetAuth(pdu.AuthTypeNTLMSSP, pdu.AuthLevelCall, creds); err != nil {
		t.Fatalf("SetAuth(call): %v", err)
	}
	if c.authLevel != pdu.AuthLevelPkt {
		t.Errorf("auth level = %d, want PKT (%d) after CALL promotion", c.authLevel, pdu.AuthLevelPkt)
	}
}

// bindAckWithChallenge returns a bind_ack whose auth verifier carries an NTLM
// CHALLENGE with the given negotiate flags.
func bindAckWithChallenge(t *testing.T, negFlags flags.NegotiateFlags) []byte {
	t.Helper()
	ch := &challenge.ChallengeMessage{NegotiateFlags: negFlags, TargetName: []byte("T\x00M\x00P\x00")}
	copy(ch.ServerChallenge[:], []byte{1, 2, 3, 4, 5, 6, 7, 8})
	chBytes, err := ch.Marshal()
	if err != nil {
		t.Fatalf("marshal CHALLENGE: %v", err)
	}
	st := pdu.SecTrailer{AuthType: pdu.AuthTypeNTLMSSP, AuthLevel: pdu.AuthLevelConnect}
	raw := append(bindAckBytes(t, 4280, 4280), st.Marshal()...)
	raw = append(raw, chBytes...)
	binary.LittleEndian.PutUint16(raw[8:10], uint16(len(raw)))
	binary.LittleEndian.PutUint16(raw[10:12], uint16(len(chBytes)))
	return raw
}

// TestCompleteAuthWithoutESS checks the NTLM bind against a server whose CHALLENGE
// does not offer extended session security. The AUTHENTICATE no longer asserts it,
// so a per-PDU level, which needs extended session security signing, fails at the
// bind instead of signing under keys the server does not use; CONNECT, which
// protects no PDU, still completes.
func TestCompleteAuthWithoutESS(t *testing.T) {
	creds, err := credentials.NewCredentials("TMP", "user", "pass", "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}
	noESS := flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_NTLM |
		flags.NTLMSSP_NEGOTIATE_SIGN | flags.NTLMSSP_NEGOTIATE_SEAL |
		flags.NTLMSSP_NEGOTIATE_128 | flags.NTLMSSP_NEGOTIATE_KEY_EXCH

	cases := []struct {
		name     string
		level    uint8
		negFlags flags.NegotiateFlags
		wantErr  bool
	}{
		{"connect without ESS", pdu.AuthLevelConnect, noESS, false},
		{"integrity without ESS", pdu.AuthLevelPktIntegrity, noESS, true},
		{"integrity with ESS", pdu.AuthLevelPktIntegrity, noESS | flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
			c := NewClient(ft)
			if err := c.SetAuth(pdu.AuthTypeNTLMSSP, tc.level, creds); err != nil {
				t.Fatalf("SetAuth: %v", err)
			}
			_, err := c.completeAuth(bindAckWithChallenge(t, tc.negFlags))
			if (err != nil) != tc.wantErr {
				t.Fatalf("completeAuth error = %v, wantErr %v", err, tc.wantErr)
			}
			if len(ft.sent) != 1 {
				t.Fatalf("sent %d PDUs, want the auth3", len(ft.sent))
			}
			_, authValue, err := pdu.ExtractAuthVerifier(ft.sent[0])
			if err != nil {
				t.Fatalf("read auth3 verifier: %v", err)
			}
			var auth authenticate.AuthenticateMessage
			if _, err := auth.Unmarshal(authValue); err != nil {
				t.Fatalf("parse AUTHENTICATE: %v", err)
			}
			offered := tc.negFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY)
			if got := auth.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY); got != offered {
				t.Errorf("AUTHENTICATE asserts extended session security = %v, want %v", got, offered)
			}
		})
	}
}
