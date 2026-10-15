// Package ibmi is the IBM i Client Access host-server network protocol
// boundary of Manticore.
//
// It owns service discovery (TCP/449 service mapper), the common 20-byte
// host-server frame and LL/CP fields, transport policy, authentication, and
// the Sign-on and Remote Command sessions. Client Access is an independent
// protocol stack: it is not DCE/RPC, does not use NDR, does not traverse SMB
// and is not a 5250 terminal session, so none of those packages are reused.
//
// Start with network/ibmi/client (connection facade, Run Command, Program
// Call) and ibmi/api/users (typed profile listing); see the package example
// in network/ibmi/client.
//
// IBM i platform definitions and program-call contracts live under the root
// ibmi tree. Packages here may import only the root ibmi, ibmi/ccsid and
// ibmi/rpc packages from it, never a concrete program wrapper or an ibmi/api
// feature package. Lower layers never import network/ibmi/client, and only
// _test.go files may import the network/ibmi/internal/hosttest fake servers.
// These rules are enforced by ibmi/architecture_test.go.
//
// Examples alias this package as ibminet, since the platform package is also
// named ibmi.
//
// # Compatibility
//
// What is supported, and how each point is established:
//
//   - Go: the version declared by go.mod; CI runs linux and windows on
//     amd64 and 386, linux on arm64, and the race detector.
//   - IBM i release: V5R4M0, validated live (Sign-on, Remote Command,
//     DSPJOB, AUTU0150 user list), both with an administrator profile and
//     with a least-privileged *USER profile (no special authorities,
//     limited capabilities): the latter gets the 0x8001 "limited user"
//     warning and lists only the profiles it may read. Request layouts follow
//     the V5R4-era JTOpen 4.8 generation (Sign-on data-stream level 2, client
//     CCSID 13488). Later releases are expected to work from IBM's documented
//     layouts but have not been validated live.
//   - Password levels: 0/1 (DES substitute) and 2/3 (SHA-1 substitute),
//     both checked against JTOpen 4.8 known-answer vectors; the live V5R4
//     system used level 0. Level 4 (PBKDF2-HMAC-SHA512 token, 64-byte
//     SHA-512 substitute, credential type 0x07) is checked against
//     known-answer vectors from a current JTOpen release but has not been
//     validated against a live level-4 system (issue #1461). Levels above 4
//     are rejected before any credential is sent. Tokens, Kerberos and
//     additional factors are not supported (issue #1456).
//   - Remote Command data-stream level: validated live at level 7.
//     Input/output parameters use usage 13 below level 5 and 33 from level
//     5; "all messages" is honored from level 7 and downgraded below it;
//     the explicit command CCSID form is used from level 10.
//   - CCSIDs: 37 and 1140 for both the system and job CCSID. Any other
//     value fails the connection unless client.Config.JobCCSIDOverride names
//     a supported one, which always adds a warning.
//   - Transport: verified TLS (TLS 1.2 or later unless
//     hostserver.TLSOptions.AllowLegacy, never without certificate and name
//     verification, never falling back to cleartext) or explicit cleartext
//     for isolated labs. Only cleartext has been exercised live; see
//     "TLS on V5R4" below.
//   - Program APIs: QGY/QGYOLAUS (formats AUTU0100, AUTU0150, AUTU0200 and
//     AUTU0250; AUTU0250 has no V5R4-era evidence), QGY/QGYGTLE and
//     QGY/QGYCLST, synchronous lists only (issue #1455). Generic Program
//     Call can carry other programs, but no other API is claimed.
//   - Host servers: Sign-on and Remote Command only (issue #1457); one
//     server job per client, no pooling (issue #1458).
//
// # TLS on V5R4
//
// Probed on the validated V5R4M0 system with Go 1.24 (2026-10-15): the
// service mapper returns ports for as-signon-s and as-rmtcmd-s, but every
// secure host-server port (9470-9476) refuses connections, as do 992 and
// 2010; port 448 accepts TCP and closes on the TLS ClientHello. No listener
// completed a handshake, so direct TLS is unavailable as that system is
// configured, and Connect with TransportTLSRequired fails without falling
// back to cleartext.
//
// Making TLS work on V5R4 also depends on the Go toolchain, which this
// library never adjusts through GODEBUG:
//
//   - SSLv3 is not supported by crypto/tls; TLS 1.0 and 1.1 need
//     TLSOptions.AllowLegacy with an explicit MinVersion.
//   - Since Go 1.24, certificates signed with SHA-1 cannot be verified at
//     all, so the host servers need a certificate with a SHA-256 (or
//     stronger) signature from a CA passed in TLSOptions.RootCAs.
//   - Legacy suites such as 3DES or RSA key exchange must be listed in
//     TLSOptions.CipherSuites with AllowLegacy.
//
// Where that is not possible, use a separately administered TLS gateway in
// front of the host servers, or explicit cleartext on an isolated network.
//
// # Safety contract
//
// Run Command executes raw CL: it is a privileged operation and caller
// text inside it is CL injection, so prefer a typed program-call API. A
// request whose reply is lost reports an uncertain outcome
// (rmtcmd.ErrUncertainOutcome) and is never retried or replayed; decide in
// application code, after creating a new client, whether repeating it is
// safe. Authentication failures are never retried because they can disable
// the profile. Returned buffers are copies, and errors, traces and
// formatting never contain credentials, seeds, substitutes, command text or
// message text.
package ibmi
