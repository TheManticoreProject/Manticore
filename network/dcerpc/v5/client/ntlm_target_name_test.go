package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// TestSetNTLMTargetNameReachesTheBlob checks the SPN set with SetNTLMTargetName is
// the MsvAvTargetName in the AUTHENTICATE's NTLMv2 blob, for both the password and
// the pass-the-hash paths, and that the default stays cifs/<DnsComputerName>.
func TestSetNTLMTargetNameReachesTheBlob(t *testing.T) {
	info, err := targetinfo.BuildServerTargetInfo("DC01", "TMP", "dc01.tmp.local", "tmp.local", nil)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo: %v", err)
	}
	chal := &challenge.ChallengeMessage{NegotiateFlags: rpcNTLMNegotiateFlags, TargetInfo: info}

	password, _ := credentials.NewCredentials("TMP", "user", "pass", "")
	hash, _ := credentials.NewCredentials("TMP", "user", "", ":8846f7eaee8fb117ad06bdd830b7586c")

	for _, tc := range []struct {
		name  string
		creds *credentials.Credentials
		spn   string
		want  string
	}{
		{"password default", password, "", "cifs/dc01.tmp.local"},
		{"password host", password, "host/dc01.tmp.local", "host/dc01.tmp.local"},
		{"hash host", hash, "host/dc01.tmp.local", "host/dc01.tmp.local"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(nil)
			c.creds = tc.creds
			c.SetNTLMTargetName(tc.spn)
			auth, err := c.buildAuthenticate(chal)
			if err != nil {
				t.Fatalf("buildAuthenticate: %v", err)
			}
			pairs, err := targetinfo.ParseTargetInfo(auth.NtChallengeResponse[44:])
			if err != nil {
				t.Fatalf("parse blob AV_PAIRs: %v", err)
			}
			if got := utf16.DecodeUTF16LE(pairs[avpair.MsvAvTargetName]); got != tc.want {
				t.Fatalf("MsvAvTargetName = %q, want %q", got, tc.want)
			}
		})
	}
}
