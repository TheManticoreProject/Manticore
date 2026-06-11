package spnego

import (
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// ServerIdentity is the server identity advertised in the NTLM CHALLENGE: the
// NetBIOS and DNS computer/domain names from the TargetInfo AV pairs, and the
// operating-system version from the CHALLENGE Version field. Individual fields are
// empty (or zero) when the server did not advertise them.
type ServerIdentity struct {
	NetBIOSComputerName string
	NetBIOSDomainName   string
	DNSComputerName     string
	DNSDomainName       string
	OSVersionMajor      uint8
	OSVersionMinor      uint8
	OSVersionBuild      uint16
}

// ServerIdentity extracts the server identity from the NTLM CHALLENGE processed
// during authentication. ok is false when no NTLM challenge has been processed yet
// (for example before CreateAuthenticateTokenFromChallengeToken is called, or for a
// non-NTLM exchange).
func (ctx *AuthContext) ServerIdentity() (ServerIdentity, bool) {
	ch := ctx.NTLMChallenge
	if ch == nil {
		return ServerIdentity{}, false
	}

	var id ServerIdentity
	// TargetInfo AV pairs carry the NetBIOS/DNS names as (non-NUL-terminated)
	// UTF-16LE; a missing pair yields a nil slice, which decodes to "".
	if pairs, err := targetinfo.ParseTargetInfo(ch.TargetInfo); err == nil {
		id.NetBIOSComputerName = utf16.DecodeUTF16LE(pairs[avpair.MsvAvNbComputerName])
		id.NetBIOSDomainName = utf16.DecodeUTF16LE(pairs[avpair.MsvAvNbDomainName])
		id.DNSComputerName = utf16.DecodeUTF16LE(pairs[avpair.MsvAvDnsComputerName])
		id.DNSDomainName = utf16.DecodeUTF16LE(pairs[avpair.MsvAvDnsDomainName])
	}
	if ch.Version != nil {
		id.OSVersionMajor = ch.Version.ProductMajorVersion
		id.OSVersionMinor = ch.Version.ProductMinorVersion
		id.OSVersionBuild = ch.Version.ProductBuild
	}
	return id, true
}
