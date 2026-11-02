package ldap

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// ldapNTLMNegotiateFlags are the flags proposed in the NTLM NEGOTIATE of an LDAP
// NTLM bind. Signing and sealing are not requested: the connection installs no NTLM
// security layer after the bind, and a DC that negotiated them would expect every
// later PDU to be protected. The CHALLENGE then offers neither, so neither is
// negotiated whatever the AUTHENTICATE asserts.
const ldapNTLMNegotiateFlags = flags.NTLMSSP_NEGOTIATE_UNICODE |
	flags.NTLMSSP_REQUEST_TARGET |
	flags.NTLMSSP_NEGOTIATE_NTLM |
	flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
	flags.NTLMSSP_NEGOTIATE_TARGET_INFO |
	flags.NTLMSSP_NEGOTIATE_128 |
	flags.NTLMSSP_NEGOTIATE_56

// nativeNTLMNegotiator implements go-ldap's NTLMNegotiator with the native NTLM
// client, so an LDAP NTLM bind carries the SPN of the LDAP service and, over
// LDAPS, the tls-server-end-point channel bindings a DC enforcing LDAP channel
// binding requires.
type nativeNTLMNegotiator struct {
	domain string

	// spnHost is the host the "ldap/<host>" SPN is built from. When it is empty
	// or an IP address, the DNS computer name from the CHALLENGE is used.
	spnHost string

	// channelBindings is the gss_channel_bindings_struct application data, or
	// nil when the connection has no TLS channel to bind to.
	channelBindings []byte

	// negotiateMessage is the NEGOTIATE as sent, kept for the AUTHENTICATE MIC.
	negotiateMessage []byte
}

// Negotiate returns the NTLM NEGOTIATE message that opens the bind.
func (n *nativeNTLMNegotiator) Negotiate(_ string, workstation string) ([]byte, error) {
	msg, err := negotiate.CreateNegotiateMessage("", workstation, ldapNTLMNegotiateFlags, nil)
	if err != nil {
		return nil, err
	}
	raw, err := msg.Marshal()
	if err != nil {
		return nil, err
	}
	n.negotiateMessage = raw
	return raw, nil
}

// ChallengeResponse answers the server's CHALLENGE with an AUTHENTICATE keyed on
// the hex-encoded NT hash go-ldap passes (it hashes a password itself).
func (n *nativeNTLMNegotiator) ChallengeResponse(challengeMessage []byte, username string, hash string) ([]byte, error) {
	chal := &challenge.ChallengeMessage{}
	if _, err := chal.Unmarshal(challengeMessage); err != nil {
		return nil, fmt.Errorf("failed to parse NTLM CHALLENGE message: %w", err)
	}

	raw, err := hex.DecodeString(hash)
	if err != nil || len(raw) != 16 {
		return nil, fmt.Errorf("NT hash must be 32 hex characters")
	}
	var ntHash [16]byte
	copy(ntHash[:], raw)

	opts := targetinfo.BlobOptions{
		TargetName:      n.targetName(chal),
		ChannelBindings: n.channelBindings,
	}
	auth, err := authenticate.CreateAuthenticateMessageWithNTHashAndOptions(chal, username, ntHash, n.domain, "", opts)
	if err != nil {
		return nil, err
	}
	// The AUTHENTICATE keeps the SIGN, SEAL and ALWAYS_SIGN flags the builder
	// asserts. The DC keys the session on the flags both sides negotiated, which
	// exclude them here, while a Server 2025 DC rejects an LDAPS AUTHENTICATE
	// that carries a MIC with all three cleared (AcceptSecurityContext data 57).
	if auth.NeedsMIC {
		if err := auth.ComputeMIC(n.negotiateMessage, challengeMessage); err != nil {
			return nil, err
		}
	}
	return auth.Marshal()
}

// targetName returns the "ldap/<host>" SPN asserted in MsvAvTargetName, or empty
// to fall back to the default when no host name is known.
func (n *nativeNTLMNegotiator) targetName(chal *challenge.ChallengeMessage) string {
	if n.spnHost != "" && net.ParseIP(n.spnHost) == nil {
		return "ldap/" + n.spnHost
	}
	if pairs, err := targetinfo.ParseTargetInfo(chal.TargetInfo); err == nil {
		if dnsComputer := pairs[avpair.MsvAvDnsComputerName]; len(dnsComputer) > 0 {
			return "ldap/" + utf16.DecodeUTF16LE(dnsComputer)
		}
	}
	return ""
}
