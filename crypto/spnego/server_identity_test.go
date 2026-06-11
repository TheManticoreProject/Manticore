package spnego_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// avPair marshals one TargetInfo AV pair (UTF-16LE value).
func avPair(t *testing.T, id avpair.AvId, value string) []byte {
	t.Helper()
	data := utf16.EncodeUTF16LE(value)
	p := avpair.AvPair{AvID: id, AvLen: uint16(len(data)), AvData: data}
	b, err := p.Marshal()
	if err != nil {
		t.Fatalf("marshal AV pair %v: %v", id, err)
	}
	return b
}

func TestAuthContextServerIdentity(t *testing.T) {
	// Build a TargetInfo blob with the four name pairs, terminated by MsvAvEOL.
	var ti []byte
	ti = append(ti, avPair(t, avpair.MsvAvNbComputerName, "SRV01")...)
	ti = append(ti, avPair(t, avpair.MsvAvNbDomainName, "CORP")...)
	ti = append(ti, avPair(t, avpair.MsvAvDnsComputerName, "srv01.corp.local")...)
	ti = append(ti, avPair(t, avpair.MsvAvDnsDomainName, "corp.local")...)
	ti = append(ti, avPair(t, avpair.MsvAvEOL, "")...)

	ctx := &spnego.AuthContext{
		NTLMChallenge: &challenge.ChallengeMessage{
			TargetInfo: ti,
			Version:    &version.Version{ProductMajorVersion: 10, ProductMinorVersion: 0, ProductBuild: 19041},
		},
	}

	id, ok := ctx.ServerIdentity()
	if !ok {
		t.Fatal("ServerIdentity ok = false, want true")
	}
	if id.NetBIOSComputerName != "SRV01" || id.NetBIOSDomainName != "CORP" {
		t.Errorf("NetBIOS names = %q / %q, want SRV01 / CORP", id.NetBIOSComputerName, id.NetBIOSDomainName)
	}
	if id.DNSComputerName != "srv01.corp.local" || id.DNSDomainName != "corp.local" {
		t.Errorf("DNS names = %q / %q, want srv01.corp.local / corp.local", id.DNSComputerName, id.DNSDomainName)
	}
	if id.OSVersionMajor != 10 || id.OSVersionMinor != 0 || id.OSVersionBuild != 19041 {
		t.Errorf("OS version = %d.%d.%d, want 10.0.19041", id.OSVersionMajor, id.OSVersionMinor, id.OSVersionBuild)
	}
}

func TestAuthContextServerIdentityNoChallenge(t *testing.T) {
	ctx := &spnego.AuthContext{}
	if _, ok := ctx.ServerIdentity(); ok {
		t.Error("ServerIdentity ok = true with no challenge, want false")
	}
}
