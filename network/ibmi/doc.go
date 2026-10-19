// Package ibmi is the IBM i Client Access host-server network protocol
// boundary of Manticore.
//
// It owns service discovery (TCP/449 service mapper), the common 20-byte
// host-server frame and LL/CP fields, transport policy, authentication, and
// the Sign-on, Remote Command, Data Queue and File (IFS) sessions. Client
// Access is an independent protocol stack: it is not DCE/RPC, does not use
// NDR, does not traverse SMB and is not a 5250 terminal session, so none of
// those packages are reused.
//
// Start with network/ibmi/client (connection facade, Run Command, Program
// Call, data queues through network/ibmi/dataqueue, IFS files through
// network/ibmi/ifs) and ibmi/api/users (typed profile listing); see the
// package example in network/ibmi/client.
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
//     CCSID 13488). V7R5M0 is also validated live with an administrator
//     profile (the same sequence plus profile tokens and data queues).
//     Releases between them are expected to work but have not been
//     validated live.
//   - Password levels: 0/1 (DES substitute) and 2/3 (SHA-1 substitute),
//     both checked against JTOpen 4.8 known-answer vectors. Level 4
//     (PBKDF2-HMAC-SHA512 token, 64-byte SHA-512 substitute, credential type
//     0x07) is checked against known-answer vectors from a current JTOpen
//     release. All three families are validated live: level 0 on V5R4, and
//     levels 0, 2 and 4 on V7R5M0, where the whole live suite passes with an
//     administrator and with a least-privileged profile (2026-10-19). Levels
//     above 4 are rejected before any password credential is sent.
//     Operators moving a system to level 4 must re-set every password while
//     it runs at level 2 or 3: a password set at level 0 or 1 has no level-4
//     hash, and Sign-on then fails with 0x00030010 (password is *NONE).
//   - Profile tokens: issued by the Sign-on server (0x7007) from a password
//     and used for Sign-on and Remote Command as credential kind 0x02,
//     validated live on V5R4 for single-use and multiple-use tokens.
//     Kerberos, identity tokens and EIM-issued tokens are not supported
//     (issue #1464), nor are additional authentication factors (issue
//     #1465).
//   - Remote Command data-stream level: validated live at level 7 (V5R4)
//     and 11 (V7R5). Input/output parameters use usage 13 below level 5 and
//     33 from level 5; "all messages" is honored from level 7 and
//     downgraded below it. From level 10 commands are sent as UTF-16BE
//     tagged CCSID 1200 (the only tagging V7R5 accepts) and messages come
//     back as 0x1106 (level 10) or 0x1107 (level 11+, IBM's RCVM0300
//     order) records, which also carry help text; level 10 itself has no
//     live target, so its 0x1106 records were captured from V7R5.
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
//     QGY/QGYCLST. Lists are built synchronously by default or, on request,
//     asynchronously in a server job with bounded polling; both modes are
//     validated live on V5R4. Generic Program Call can carry other
//     programs, but no other API is claimed.
//   - Host servers: Sign-on, Remote Command, Data Queue and File (read
//     only); Database and Network Print are not implemented (issue #1457).
//     One server job per session, no pooling (issue #1458).
//   - Data Queue: every operation (attributes, create, write, read, peek,
//     keyed searches, clear, delete) validated live on V5R4 with the
//     least-privileged profile, on queues the test creates and deletes.
//     That system refuses empty entries with CPF2498; the sender
//     information layout (job name, job user, job number, current user)
//     was confirmed there.
//   - File (IFS): open, read and close of a stream file validated live on
//     V5R4 with the least-privileged profile (a 3.5 MB file read whole,
//     end of file, missing file and directory errors) and on V7R5M0 with
//     an administrator profile; requests always use the level-8 layouts,
//     which V7R5 accepts although it reports level 24. Writing, directory
//     listing and attribute queries are not implemented.
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
