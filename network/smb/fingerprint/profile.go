// Package fingerprint holds the externally observable constants this SMB
// implementation emits, in one place, so that what a peer sees is decided
// deliberately rather than by whichever literal happened to be typed at each
// call site.
//
// Every field here reaches the wire. Before this package they were scattered
// across the client, the facade and the NetBIOS transport, and several of them
// announced this project by name — a session setup carrying "Manticore" as its
// native operating system identifies the implementation to anyone reading the
// exchange, before any behavioural analysis is needed.
//
// # Provenance
//
// A profile is only worth as much as the evidence behind its values, so each
// carries a Provenance describing where its values come from. Three kinds
// appear here:
//
//   - measured from a real peer, which is the strongest;
//   - required by a specification, which is next; and
//   - chosen by this implementation, which is honest but says nothing about
//     resembling anything else.
//
// There is deliberately no profile purporting to reproduce a Windows *client*.
// Doing that needs a capture of a Windows client talking to a server, which
// this repository does not have, and inventing plausible-looking values would
// be worse than having none: the wire comparison harness would then pin
// fabricated bytes and report agreement with something that never existed.
package fingerprint

import (
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

// Profile is one coherent set of the values this implementation puts on the
// wire. Fields are grouped by the exchange they appear in.
type Profile struct {
	// Name identifies the profile, and Provenance says where its values come
	// from. Provenance is not decoration: a value nobody can source is a value
	// nobody should pin a test to.
	Name       string
	Provenance string

	// NativeOS and NativeLanMan are the informational strings carried in an SMB
	// 1.0 SESSION_SETUP_ANDX ([MS-CIFS] 2.2.4.53.2). No behaviour depends on
	// them, which is exactly why they are worth choosing deliberately: they are
	// the most direct description of an implementation available to an observer.
	NativeOS     string
	NativeLanMan string

	// NTLMVersion is the VERSION structure carried in NTLM messages
	// ([MS-NLMP] 2.2.2.10). A build number of zero identifies no real Windows
	// release, so a profile that sets this at all should set it to something a
	// peer could plausibly be running.
	NTLMVersion version.Version

	// CallingNameFallback is the NetBIOS CALLING name used when the local
	// hostname cannot be determined (RFC 1002 4.3.2). It reaches the wire on
	// port 139 before anything else does.
	CallingNameFallback string

	// Dialects are the SMB2 dialect revisions offered, in ascending order as
	// [MS-SMB2] 2.2.3 requires.
	Dialects []dialects.Dialect

	// Capabilities and SecurityMode are the SMB2 NEGOTIATE fields of the same
	// name.
	Capabilities capabilities.Capabilities
	SecurityMode securitymode.SecurityMode

	// Ciphers are the SMB 3.1.1 encryption ciphers offered, in preference order.
	Ciphers []uint16

	// PreauthSaltLength is the size of the salt in the SMB 3.1.1 pre-auth
	// integrity negotiate context.
	PreauthSaltLength int
}

// Clone returns a copy whose slices are independent, so a caller adjusting one
// field of a profile cannot alter the package-level profile it started from.
func (p *Profile) Clone() *Profile {
	if p == nil {
		return nil
	}
	out := *p
	out.Dialects = append([]dialects.Dialect(nil), p.Dialects...)
	out.Ciphers = append([]uint16(nil), p.Ciphers...)
	return &out
}
