package client

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// TestCompleteAuthComputesMIC drives the NTLM bind against the repository's acceptor
// with a CHALLENGE that carries an MsvAvTimestamp. The AUTHENTICATE then announces a
// MIC in its MsvAvFlags, so it MUST carry a valid one over NEGOTIATE||CHALLENGE||
// AUTHENTICATE (MS-NLMP 3.1.5.1.2); a zeroed MIC fails the acceptor's check.
func TestCompleteAuthComputesMIC(t *testing.T) {
	const domain, username, password = "CORP", "alice", "Passw0rd!"
	creds, err := credentials.NewCredentials(domain, username, password, "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}

	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	c := NewClient(ft)
	c.workstation = "WS"
	if err := c.SetAuth(pdu.AuthTypeNTLMSSP, pdu.AuthLevelPktPrivacy, creds); err != nil {
		t.Fatalf("SetAuth: %v", err)
	}
	negToken, err := c.negotiateToken()
	if err != nil {
		t.Fatalf("negotiateToken: %v", err)
	}

	timestamp := []byte{0x00, 0x80, 0x3e, 0xd5, 0xde, 0xb1, 0x9d, 0x01}
	info, err := targetinfo.BuildServerTargetInfo("SERVER01", "CORP", "server01.corp.local", "corp.local", timestamp)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo: %v", err)
	}
	acc := spnego.NewAcceptContext("CORP", challenge.TargetTypeDomain, info)
	acc.CredentialLookup = func(gotDomain, gotUsername string) ([16]byte, bool) {
		if !strings.EqualFold(gotDomain, domain) || !strings.EqualFold(gotUsername, username) {
			return [16]byte{}, false
		}
		return nt.NTHash(password), true
	}
	if _, err := acc.AcceptNegotiateToken(negToken); err != nil {
		t.Fatalf("AcceptNegotiateToken: %v", err)
	}

	// Carry the acceptor's CHALLENGE, byte for byte, in the bind_ack auth verifier.
	chBytes := acc.ChallengeMessageBytes
	st := pdu.SecTrailer{AuthType: pdu.AuthTypeNTLMSSP, AuthLevel: pdu.AuthLevelPktPrivacy}
	ack := append(bindAckBytes(t, 4280, 4280), st.Marshal()...)
	ack = append(ack, chBytes...)
	binary.LittleEndian.PutUint16(ack[8:10], uint16(len(ack)))
	binary.LittleEndian.PutUint16(ack[10:12], uint16(len(chBytes)))

	if _, err := c.completeAuth(ack); err != nil {
		t.Fatalf("completeAuth: %v", err)
	}
	if len(ft.sent) != 1 {
		t.Fatalf("sent %d PDUs, want the auth3", len(ft.sent))
	}
	_, authValue, err := pdu.ExtractAuthVerifier(ft.sent[0])
	if err != nil {
		t.Fatalf("read auth3 verifier: %v", err)
	}
	if err := acc.AcceptAuthenticateToken(authValue); err != nil {
		t.Fatalf("AcceptAuthenticateToken: %v", err)
	}
	if !acc.Authenticate.AnnouncesMIC() {
		t.Fatal("AUTHENTICATE does not announce a MIC although the CHALLENGE carried a timestamp")
	}
	if err := acc.Verify(); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}
