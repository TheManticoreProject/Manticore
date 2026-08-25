// Package server implements the server side of the SMB 1.0 (CIFS) protocol.
//
// It is a library, not a tool: it exposes a Server that listens, decodes
// requests with the message layer in network/smb/smb_v10/message, and answers
// them, plus a Handler chain a caller can use to observe or intercept requests
// before the built-in dispatch sees them. The shape mirrors
// network/llmnr/server, so the two compose: a name-service poisoner can steer a
// client at this server.
//
// # What is implemented
//
// This package is being built up in phases, and it is deliberately explicit
// about where it currently stops, because a partial SMB server answers enough of
// the protocol to look functional while refusing everything that matters.
//
// Implemented:
//   - Listening on Direct TCP (445) and NetBIOS over TCP (139), via
//     network/smb/common/transport.
//   - The per-connection receive loop, request decoding, the handler chain, and
//     response framing with correlated reply headers.
//   - Error responses in both encodings: the NTSTATUS form, and the legacy
//     SMBSTATUS class/code form for a client that did not negotiate
//     SMB_FLAGS2_NT_STATUS_ERROR_CODES.
//   - SMB_COM_NEGOTIATE, selecting the NT LM 0.12 dialect under extended
//     security.
//   - SMB_COM_SESSION_SETUP_ANDX, including verifying the response against a
//     credential, establishing a session, and the guest and anonymous policies.
//   - SMB_COM_LOGOFF_ANDX.
//   - SMB_COM_ECHO.
//   - Message signing in both directions, when the policy calls for it.
//   - Tree connect and disconnect against a registered share.
//   - File service: open and create, read, write, close, flush, delete, rename,
//     and the directory create, remove and check commands.
//   - Directory enumeration and the information levels, over TRANSACTION2:
//     FIND_FIRST2 and FIND_NEXT2 with search handles, the query and set levels
//     for a path and for an open handle, and the volume levels. Requests and
//     responses both fragment across as many messages as they need.
//   - Security descriptors and file-system controls, over NT_TRANSACT:
//     QUERY_SECURITY_DESC, SET_SECURITY_DESC and IOCTL. SMB_COM_NT_CANCEL is
//     accepted silently, since nothing here leaves a request outstanding.
//   - Named pipes, over TRANSACTION: a pipe is opened on a pipe share like a
//     file, and TRANS_TRANSACT_NMPIPE writes a message to the handle and returns
//     the answer. That write-then-read is the operation MS-RPC travels over, so a
//     PipeHandler is all an RPC service needs to be reachable over SMB1.
//   - The volume queries a client actually asks: the TRANSACTION2 volume levels,
//     the pass-through information classes above 0x03E8 that carry the native
//     ones, and the legacy SMB_COM_QUERY_INFORMATION_DISK. A client asks about
//     free space after a listing whether or not anything wanted it, so leaving
//     these unanswered puts an error in every session.
//
// All three transaction families share one reassembly, since they are the same
// shape at different field widths: totals, a per-message count and a
// displacement, with the subcommand selected by a setup word, a Function field or
// a name.
//
// Not yet implemented, and answered with STATUS_NOT_IMPLEMENTED: byte-range
// locking, seek, the legacy SMB_COM_OPEN_ANDX, and batched AndX chains beyond
// their first command.
//
// NT_TRANSACT_NOTIFY_CHANGE is deliberately absent rather than pending. It needs
// two things this package does not have: a FileSystem that can be watched, and a
// connection whose write path can be used from outside the request that is being
// served — a notification is answered when the change happens, not when it is
// asked for. Both are architectural additions, and half of either would be worse
// than the honest refusal.
//
// NT_TRANSACT_CREATE and NT_TRANSACT_RENAME are also absent by choice: they
// duplicate SMB_COM_NT_CREATE_ANDX and SMB_COM_RENAME, which are served. The
// quota subcommands are absent because nothing here tracks a quota, and a number
// invented for them is a number a client would believe.
//
// # Shares
//
// A Share is registered with AddShare and backed by a FileSystem.
// NewLocalFileSystem serves a directory on the host; NewMemoryFileSystem serves
// storage that never touches disk, which is what the tests use and what a share
// meant to look real without being real would use.
//
// A share may be marked ReadOnly, which refuses every modifying command whatever
// access the client asked for. That is enforced in the handlers rather than left
// to the backend, so a backend cannot forget it.
//
// # Security descriptors
//
// A Share may carry a SecurityProvider, which answers the NT_TRANSACT security
// subcommands. NewReflectiveSecurityProvider derives a descriptor from the
// share's own configuration: a read-only share does not describe write access,
// because it does not grant any. That is the point of deriving one rather than
// returning a fixed descriptor — a client uses a descriptor to predict what it
// will be allowed to do, so one that disagreed with the handlers would make the
// client wrong. For the same reason it refuses a change instead of accepting one
// it has nowhere to store.
//
// A share with no provider answers STATUS_NOT_SUPPORTED rather than inventing a
// descriptor.
//
// # Named pipes
//
// A Share of type ShareTypeNamedPipe carries a PipeHandler instead of a
// FileSystem. A client opens a pipe on it with SMB_COM_NT_CREATE_ANDX and then
// transacts on the handle: [MS-CIFS] identifies the pipe a transaction acts on by
// the FID in the request's setup words, not by the name the request carries, so
// the handle is what matters and the name is boilerplate.
//
// An answer larger than the client's buffer is cut to fit and reported with
// STATUS_BUFFER_OVERFLOW, which is what tells the client to read again. Reporting
// plain success would leave an RPC client parsing a truncated response as a whole
// one.
//
// # Character encoding
//
// Unicode is a per-message property, not a per-connection one: SMB_FLAGS2_UNICODE
// is set on each message, and a client may negotiate Unicode and then send a
// request in OEM. So every name is read and written in the encoding that message
// declared, never in the connection's.
//
// The consequences are easy to underestimate. A null-terminated field ends at its
// first null CHARACTER, so a Unicode name has a two-byte terminator and a
// single-byte scan ends it after one character. A Unicode field also has to begin
// on a 2-byte boundary measured from the start of the SMB header, so a padding byte
// stands before it whenever the fields ahead of it did not leave it aligned — which
// for the second name of a rename depends on the length of the first. And a name in
// a response is read by the client as whatever the message declared, so a name
// written in the other encoding produces a reply of the right shape and the wrong
// text rather than an error.
//
// # Path containment
//
// Every path a client sends passes through resolvePath before any backend sees
// it, and a backend is entitled to assume the result cannot escape the share. The
// resolver refuses rather than normalises: a path containing ".." is rejected
// outright instead of being rewritten, because rewriting turns a traversal attempt
// into a successful access somewhere unintended.
//
// LocalFileSystem adds a second, independent check, because path validation
// cannot see a symbolic link inside the share pointing out of it: every resolved
// host path is compared against the share root again after the host has followed
// its links.
//
// # Authentication
//
// Config.Authenticator resolves a claimed identity to its NT hash, and
// StaticAccounts builds one from a fixed list. With no Authenticator no logon
// can succeed, which is the configuration a server whose purpose is harvesting
// responses wants.
//
// Config.AllowGuest admits an identity the store does not know, reporting
// SMB_SETUP_GUEST so the client knows it was not authenticated as itself, and
// Config.AllowAnonymous admits a null session. Neither derives a key, so neither
// can sign: under a policy that requires signatures they are refused outright
// rather than granted a session that could not carry a single request.
//
// # Signing
//
// Config.SigningPolicy selects whether signatures are unsupported, offered or
// demanded, and only what the server will honour is advertised. Signing is
// bootstrapped by the authentication exchange itself: the client signs its
// AUTHENTICATE with the key it derived, and the server can only check that once
// it has derived the same key from the response. From then on every request must
// carry a valid signature at the number the exchange has reached, and every
// response is signed at the number above.
//
// # Credential capture
//
// A CaptureHandler registered on the server harvests the NTLM response from
// every attempt and renders it in hashcat form, so material a server cannot
// verify can be cracked offline instead. It composes with the above: a server
// with no Authenticator refuses every logon and captures every response, while
// one with an Authenticator serves the identities it knows and captures the rest.
//
// # Security posture
//
// The receive loop is the attack surface of a listening service, so it is
// written to survive arbitrary input: a frame that is not an SMB message, or
// whose header is well formed but whose body will not decode, is answered or
// dropped rather than propagated, and a panic in a handler takes down only that
// connection. FuzzServerFrame in this package exercises that path.
//
// A handle is not always backed by a file — a pipe handle has a handler instead,
// and a backend may decline to open a directory — so every command that reads or
// writes through one checks. The client chooses the handle, so an unguarded
// dereference there is reachable by anyone who can open a pipe.
//
// # Interoperability
//
// The unit suite pairs this server with the SMB1 client in this repository, which
// is fast to work with but shares this implementation's assumptions: a wire detail
// both halves get wrong agrees with itself, and every round-trip passes.
// live_interop_integration_test.go exists for that reason. Behind the
// "integration" build tag, it drives a third-party client and a third-party RPC
// client against a server started in-process, and asserts a clean session: a
// listing by name, a file in both directions, the name-carrying commands, a
// mandatory-signing session verified in both directions, and an RPC bind completed
// over a named pipe.
//
// Not covered there: the NT_TRANSACT security-descriptor path, because the tool
// that would drive it cannot be pointed at a non-privileged port. It is covered by
// unit tests that parse the descriptor back with an independent parser.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
package server
