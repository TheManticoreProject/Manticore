package ldap

import (
	"bytes"
	"crypto/md5"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
)

// TestNativeNTLMNegotiatorAssertsSPNAndChannelBindings runs the LDAP negotiator
// against an acceptor-built CHALLENGE and checks the AUTHENTICATE asserts the
// ldap/ SPN and the tls-server-end-point channel bindings. The expected bindings
// value is the MD5 of the structure the Kerberos GSSAPI bind already sends to
// the same DCs, so the two mechanisms bind to the TLS channel identically.
func TestNativeNTLMNegotiatorAssertsSPNAndChannelBindings(t *testing.T) {
	cbt := append([]byte(tlsServerEndPointPrefix), bytes.Repeat([]byte{0x5a}, 32)...)
	wantCB := md5.Sum(gssapi.GSSChannelBindings(cbt))

	for _, tc := range []struct {
		name    string
		spnHost string
		want    string
	}{
		{"IP host falls back to the DNS computer name", "10.0.0.1", "ldap/dc01.tmp.local"},
		{"empty host falls back to the DNS computer name", "", "ldap/dc01.tmp.local"},
		{"FQDN host", "dc01-alias.tmp.local", "ldap/dc01-alias.tmp.local"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := &nativeNTLMNegotiator{domain: "TMP", spnHost: tc.spnHost, channelBindings: cbt}
			negBytes, err := n.Negotiate("TMP", "")
			if err != nil {
				t.Fatalf("Negotiate: %v", err)
			}
			neg := &negotiate.NegotiateMessage{}
			if _, err := neg.Unmarshal(negBytes); err != nil {
				t.Fatalf("parse NEGOTIATE: %v", err)
			}
			if neg.NegotiateFlags&(flags.NTLMSSP_NEGOTIATE_SIGN|flags.NTLMSSP_NEGOTIATE_SEAL) != 0 {
				t.Fatalf("NEGOTIATE requests signing or sealing (flags 0x%08x), which the LDAP connection does not apply", uint32(neg.NegotiateFlags))
			}

			info, err := targetinfo.BuildServerTargetInfo("DC01", "TMP", "dc01.tmp.local", "tmp.local", make([]byte, 8))
			if err != nil {
				t.Fatalf("BuildServerTargetInfo: %v", err)
			}
			chal, err := challenge.CreateChallengeMessage(neg, [8]byte{1, 2, 3, 4, 5, 6, 7, 8}, "TMP", challenge.TargetTypeDomain, info, nil)
			if err != nil {
				t.Fatalf("CreateChallengeMessage: %v", err)
			}
			chalBytes, err := chal.Marshal()
			if err != nil {
				t.Fatalf("marshal CHALLENGE: %v", err)
			}

			authBytes, err := n.ChallengeResponse(chalBytes, "user", "8846f7eaee8fb117ad06bdd830b7586c")
			if err != nil {
				t.Fatalf("ChallengeResponse: %v", err)
			}
			auth := &authenticate.AuthenticateMessage{}
			if _, err := auth.Unmarshal(authBytes); err != nil {
				t.Fatalf("parse AUTHENTICATE: %v", err)
			}
			if chal.NegotiateFlags&(flags.NTLMSSP_NEGOTIATE_SIGN|flags.NTLMSSP_NEGOTIATE_SEAL) != 0 {
				t.Errorf("CHALLENGE offers signing or sealing (flags 0x%08x) the NEGOTIATE did not request", uint32(chal.NegotiateFlags))
			}
			if auth.MIC == [16]byte{} {
				t.Error("CHALLENGE carried a timestamp but the AUTHENTICATE has no MIC")
			}
			pairs, err := targetinfo.ParseTargetInfo(auth.NtChallengeResponse[44:])
			if err != nil {
				t.Fatalf("parse blob AV_PAIRs: %v", err)
			}
			if got := utf16.DecodeUTF16LE(pairs[avpair.MsvAvTargetName]); got != tc.want {
				t.Errorf("MsvAvTargetName = %q, want %q", got, tc.want)
			}
			if got := pairs[avpair.MsvAvChannelBindings]; !bytes.Equal(got, wantCB[:]) {
				t.Errorf("MsvAvChannelBindings = %x, want %x", got, wantCB)
			}
		})
	}
}

// TestNativeNTLMNegotiatorRejectsMalformedHash checks a hash that is not 16 bytes
// of hex is refused rather than keying the response on garbage.
func TestNativeNTLMNegotiatorRejectsMalformedHash(t *testing.T) {
	n := &nativeNTLMNegotiator{domain: "TMP"}
	negBytes, _ := n.Negotiate("TMP", "")
	neg := &negotiate.NegotiateMessage{}
	if _, err := neg.Unmarshal(negBytes); err != nil {
		t.Fatalf("parse NEGOTIATE: %v", err)
	}
	chal, err := challenge.CreateChallengeMessage(neg, [8]byte{}, "TMP", challenge.TargetTypeDomain, nil, nil)
	if err != nil {
		t.Fatalf("CreateChallengeMessage: %v", err)
	}
	chalBytes, _ := chal.Marshal()
	for _, hash := range []string{"", "zz", "8846f7eaee8fb117ad06bdd830b758"} {
		if _, err := n.ChallengeResponse(chalBytes, "user", hash); err == nil {
			t.Errorf("hash %q accepted", hash)
		}
	}
}
