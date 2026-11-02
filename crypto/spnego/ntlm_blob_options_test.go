package spnego

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// ntlmv2BlobAvPairs returns the AV_PAIRs of an NTLMv2 NtChallengeResponse: they
// follow the 16-byte NTProofStr and the 28-byte fixed blob header ([MS-NLMP] 2.2.2.7).
func ntlmv2BlobAvPairs(t *testing.T, ntResponse []byte) map[avpair.AvId][]byte {
	t.Helper()
	if len(ntResponse) < 44 {
		t.Fatalf("NtChallengeResponse is %d bytes, too short for an NTLMv2 response", len(ntResponse))
	}
	pairs, err := targetinfo.ParseTargetInfo(ntResponse[44:])
	if err != nil {
		t.Fatalf("parse blob AV_PAIRs: %v", err)
	}
	return pairs
}

// TestAuthContextTargetNameAndChannelBindings drives an exchange whose initiator
// sets TargetName and ChannelBindings, for both the password and the NT-hash
// paths, and checks the acceptor verifies it and receives the caller's SPN and
// the MD5 of its channel bindings rather than the cifs/ and all-zero defaults.
func TestAuthContextTargetNameAndChannelBindings(t *testing.T) {
	appData := []byte("tls-server-end-point:0123456789abcdef0123456789abcdef")
	ntHash := nt.NTHash(testPassword)

	for _, passTheHash := range []bool{false, true} {
		ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
		ctx.CredentialLookup = lookupFor(testDomain, testUsername, testPassword, nil)

		client := NewAuthContext(AuthTypeNTLM, testDomain, testUsername, testPassword, testWorkstation, true)
		if passTheHash {
			client.Password = ""
			client.NTHash = hex.EncodeToString(ntHash[:])
		}
		client.TargetName = "ldap/server01.lab.example.local"
		client.ChannelBindings = appData

		v := version.DefaultVersion()
		negotiateToken, err := client.CreateNegotiateToken(clientNegotiateFlags, &v)
		if err != nil {
			t.Fatalf("CreateNegotiateToken: %v", err)
		}
		challengeToken, err := ctx.AcceptNegotiateToken(negotiateToken)
		if err != nil {
			t.Fatalf("AcceptNegotiateToken: %v", err)
		}
		authenticateToken, err := client.CreateAuthenticateTokenFromChallengeToken(challengeToken)
		if err != nil {
			t.Fatalf("CreateAuthenticateTokenFromChallengeToken: %v", err)
		}
		if err := ctx.AcceptAuthenticateToken(authenticateToken); err != nil {
			t.Fatalf("AcceptAuthenticateToken: %v", err)
		}
		if err := ctx.Verify(); err != nil {
			t.Fatalf("passTheHash=%v: Verify: %v", passTheHash, err)
		}

		pairs := ntlmv2BlobAvPairs(t, ctx.Authenticate.NtChallengeResponse)
		if got := utf16.DecodeUTF16LE(pairs[avpair.MsvAvTargetName]); got != client.TargetName {
			t.Errorf("passTheHash=%v: MsvAvTargetName = %q, want %q", passTheHash, got, client.TargetName)
		}
		if got, want := pairs[avpair.MsvAvChannelBindings], targetinfo.ChannelBindingsHash(appData); !bytes.Equal(got, want) {
			t.Errorf("passTheHash=%v: MsvAvChannelBindings = %x, want %x", passTheHash, got, want)
		}
	}
}
