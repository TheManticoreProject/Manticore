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
//
//   - Listening on Direct TCP (445) and NetBIOS over TCP (139), via
//     network/smb/common/transport.
//
//   - The per-connection receive loop, request decoding, the handler chain, and
//     response framing with correlated reply headers.
//
//   - Error responses in both encodings: the NTSTATUS form, and the legacy
//     SMBSTATUS class/code form for a client that did not negotiate
//     SMB_FLAGS2_NT_STATUS_ERROR_CODES.
//
//   - SMB_COM_NEGOTIATE, selecting the NT LM 0.12 dialect under extended
//     security.
//
//   - SMB_COM_SESSION_SETUP_ANDX, including verifying the response against a
//     credential, establishing a session, and the guest and anonymous policies.
//
//   - SMB_COM_LOGOFF_ANDX.
//
//   - SMB_COM_ECHO.
//
//   - Message signing in both directions, when the policy calls for it.
//
//   - Tree connect and disconnect against a registered share.
//
//   - File service: open and create, read, write, close, flush, delete, rename,
//     and the directory create, remove and check commands.
//
//   - SMB_COM_NT_RENAME, which renames an entry at SMB_NT_RENAME_RENAME_FILE and
//     gives it a second name at SMB_NT_RENAME_SET_LINK_INFO. Linking needs a
//     backend that implements Linker; one that does not is answered
//     STATUS_NOT_SUPPORTED rather than given a copy, because a copy is a second
//     file and not a second name, and the two diverge as soon as either is
//     written.
//
//     SMB_COM_COPY and SMB_COM_MOVE stay refused: [MS-CIFS] sections 2.2.4.37
//     and 2.2.4.38 record both as obsolete in the NT LAN Manager dialect — the
//     only dialect this server speaks — and have servers answer
//     STATUS_NOT_IMPLEMENTED, which is what happens.
//
//   - The core-set file access commands, which a client that does not use the NT
//     commands opens and transfers with: SMB_COM_OPEN, SMB_COM_OPEN_ANDX,
//     SMB_COM_CREATE, SMB_COM_CREATE_NEW, SMB_COM_CREATE_TEMPORARY,
//     SMB_COM_READ, SMB_COM_WRITE, SMB_COM_WRITE_AND_CLOSE, SMB_COM_SEEK, the
//     deprecated SMB_COM_TREE_CONNECT, and SMB_COM_PROCESS_EXIT.
//
//     Two of these behave in ways worth knowing. A write of zero bytes sets the
//     file's length rather than writing nothing, which is what [MS-CIFS] defines
//     it as. And SMB_COM_PROCESS_EXIT releases the handles held by the PID in the
//     request header, which is why an open records the PID that made it — the
//     command exists to clean up after one failed client process on a connection
//     that may be serving several.
//
//     SMB_COM_CLOSE_AND_TREE_DISC stays refused on purpose: [MS-CIFS] section
//     2.2.4.45 records it as reserved but never implemented and has servers
//     answer STATUS_NOT_IMPLEMENTED, which is what happens.
//
//   - The core-set file information commands, which describe a file without a
//     transaction: SMB_COM_QUERY_INFORMATION and SMB_COM_SET_INFORMATION by path,
//     and SMB_COM_QUERY_INFORMATION2 and SMB_COM_SET_INFORMATION2 on a handle.
//     Their timestamps are MS-DOS date and time pairs with two-second resolution,
//     or a UTIME in seconds, so a client reading one back sees less precision than
//     the TRANSACTION2 levels carry — a property of the wire format rather than of
//     the storage.
//
//   - The core-set directory search: SMB_COM_SEARCH, SMB_COM_FIND,
//     SMB_COM_FIND_UNIQUE and SMB_COM_FIND_CLOSE. These keep no state on the
//     server — the position travels in the resume key and the directory is read
//     again each call, which is what a command with no close needs, since a
//     client that walks away mid-listing would otherwise leak an allocation. The
//     cost is that the listing is a fresh view rather than a snapshot, so an
//     entry created or removed between calls may be seen twice or missed;
//     TRANS2_FIND_FIRST2 is the level that keeps a snapshot.
//
//     An entry's name field is a fixed thirteen bytes, so a name that does not
//     fit an 8.3 field is omitted from these listings. There is no 8.3 alias to
//     substitute: truncating would name a different file, and inventing an alias
//     would hand the client a name no subsequent open could resolve. A client
//     that needs those names has to use TRANS2_FIND_FIRST2.
//
//   - The pre-NT information levels, so a client that did not negotiate
//     CAP_NT_FIND can still work: SMB_INFO_STANDARD and SMB_INFO_QUERY_EA_SIZE
//     for find and for query, SMB_INFO_IS_NAME_VALID, SMB_QUERY_FILE_STREAM_INFO,
//     SMB_QUERY_FILE_COMRESSION_INFO, and the SMB_INFO_ALLOCATION and
//     SMB_INFO_VOLUME volume levels.
//
//     Their entries are packed rather than linked: the NT levels begin each entry
//     with the offset of the next, and these begin with a date, so the buffer
//     assembly has to know which model a level uses. Writing a chain terminator
//     into a pre-NT buffer would overwrite the first entry's timestamps and
//     produce a listing that looks valid.
//
//   - Directory enumeration and the information levels, over TRANSACTION2:
//     FIND_FIRST2 and FIND_NEXT2 with search handles, the query and set levels
//     for a path and for an open handle, and the volume levels. Requests and
//     responses both fragment across as many messages as they need.
//
//   - Security descriptors and file-system controls, over NT_TRANSACT:
//     QUERY_SECURITY_DESC, SET_SECURITY_DESC and IOCTL. SMB_COM_NT_CANCEL is
//     accepted silently, since nothing here leaves a request outstanding.
//
//   - Named pipes, over TRANSACTION: a pipe is opened on a pipe share like a
//     file, and TRANS_TRANSACT_NMPIPE writes a message to the handle and returns
//     the answer. That write-then-read is the operation MS-RPC travels over, so a
//     PipeHandler is all an RPC service needs to be reachable over SMB1.
//     PipeHandler is all an RPC service needs to be reachable over SMB1. An answer
//     too large for one response is collected with TRANS_READ_NMPIPE,
//     TRANS_PEEK_NMPIPE or SMB_COM_READ_ANDX on the same handle.
//
//   - The volume queries a client actually asks: the TRANSACTION2 volume levels,
//     the pass-through information classes above 0x03E8 that carry the native
//     ones, and the legacy SMB_COM_QUERY_INFORMATION_DISK. A client asks about
//     free space after a listing whether or not anything wanted it, so leaving
//     these unanswered puts an error in every session.
//
// A read or a write may exceed the negotiated MaxBufferSize once both sides have
// agreed CAP_LARGE_READX or CAP_LARGE_WRITEX, bounded by Config.MaxLargeTransfer.
// Signing overrides that: a signed response is verified over the bytes as sent, so
// a client that sized its buffer to MaxBufferSize and received more would fail the
// signature rather than merely truncate, and [MS-SMB] has clients hold to
// MaxBufferSize whenever signing is active.
//
// The ceiling stays under 0x10000 because SMB_Data.ByteCount is a USHORT that no
// extension widens, so a larger data block could not describe its own length.
//
//   - The pass-through information classes for files, in both directions: the
//     basic, standard, internal, EA, access, position, name, alternate-name,
//     network-open and all classes for a query, and the basic, disposition,
//     allocation, end-of-file and rename classes for a set. Those structures come
//     from windows/filesystem rather than being assembled here, so their layouts
//     are the ones the rest of the repository agrees on.
//
//     A pass-through structure's strings are UTF-16LE whatever the message
//     declared, which is the one place in this package where a name's encoding
//     does not follow SMB_FLAGS2_UNICODE: an SMB level carries an SMB string, but
//     a pass-through level carries the [MS-FSCC] structure verbatim.
//
// All three transaction families share one reassembly, since they are the same
// shape at different field widths: totals, a per-message count and a
// displacement, with the subcommand selected by a setup word, a Function field or
// a name.
//
//   - Batched ("AndX") requests: every command in a chain runs, in order, and all
//     the answers return in one message. A command sees the identifiers as they
//     stand when it runs rather than as the client sent them, which is what makes
//     a session setup batched with a tree connect work — the client had no UID to
//     send. A failure ends the chain and the error response closes it, per
//     [MS-CIFS] 3.3.4.1, so the answers already produced still come back.
//   - Byte-range locking, over SMB_COM_LOCKING_ANDX: locks and unlocks in one
//     atomic request, in both range formats, exclusive and shared. Overlapping
//     locks are refused, an unlock of a range the handle does not hold is
//     refused, and closing a handle releases what it held. The locks are
//     enforced: a read or a write through another handle onto an exclusively
//     locked range is refused, and a shared lock refuses only writes. A request
//     that offers to wait waits, bounded by Config.MaxLockWait.
//
// Not yet implemented, and answered with STATUS_NOT_IMPLEMENTED: seek and the
// legacy SMB_COM_OPEN_ANDX.
//
// NT_TRANSACT_NOTIFY_CHANGE is deliberately absent rather than pending. It needs
// two things this package does not have: a FileSystem that can be watched, and a
// connection whose write path can be used from outside the request that is being
// served — a notification is answered when the change happens, not when it is
// asked for. Both are architectural additions, and half of either would be worse
// than the honest refusal.
//
// NT_TRANSACT_CREATE and TRANS2_OPEN2 are served, as are TRANS2_CREATE_DIRECTORY
// and the extended attributes those two can carry — except that the attributes
// themselves are not stored, and the responses report a length of zero, which is
// how the format says a file has none.
//
// TRANS2_SET_FS_INFORMATION and NT_TRANSACT_RENAME are reserved and were never
// implemented, and each is refused with the status its own section names rather
// than with the tables' generic STATUS_NOT_IMPLEMENTED: [MS-CIFS] 2.2.6.5 requires
// STATUS_SMB_NO_SUPPORT for the first and 2.2.7.5 requires STATUS_SMB_BAD_COMMAND
// for the second.
//
// The quota subcommands are absent because nothing here tracks a quota, and a
// number invented for them is a number a client would believe. TRANS2_FSCTL,
// TRANS2_IOCTL2 and TRANS2_SESSION_SETUP are reserved and take the generic
// refusal, which is what their sections require.
//
// # Shares
//
// A FileSystem may also implement Linker, which lets SMB_COM_NT_RENAME give a
// file a second name. It is a separate interface so that adding it does not break
// a backend outside this repository: one that cannot link simply does not
// implement it.
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
// # Byte-range locks
//
// Locks are held by the server, in a table on the Share, rather than delegated to
// the FileSystem. That is deliberate. SMB lock semantics are not the host's: a
// lock is held on a FID and excludes every other FID onto the same file, whoever
// opened it, and a MemoryFileSystem has no host locks to delegate to while a
// LocalFileSystem's would carry the platform's rules rather than the protocol's.
// Holding them here means every backend gets the same, correct semantics and none
// of them can forget to.
//
// The table belongs to the Share because a lock is a statement about a file, and
// the handles it has to exclude are on other connections as much as on the one
// that took it.
//
// A blocked request waits by polling, and Config.MaxLockWait bounds how long. A
// client may ask to wait forever, and the connection serves nothing else while it
// does, so an unbounded wait would let a client stall itself with no way out — the
// cancel it would send arrives behind the request it wanted to cancel.
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
// The part that did not fit is kept on the handle, and reading again is how the
// client collects it: SMB_COM_READ_ANDX on the pipe FID, TRANS_READ_NMPIPE, or
// TRANS_PEEK_NMPIPE to size the read first. A read once the answer is exhausted
// returns no data rather than an error, because a client reads a pipe only after
// being told more remains, so an empty read is the end of the answer rather than a
// failure. Writing a pipe handle with SMB_COM_WRITE_ANDX is still refused: a
// handler answers a transaction rather than accepting a stream, so there is
// nowhere for the write to go.
//
// The handler is asked for the whole answer rather than for the part that fits,
// bounded by maxPipeAnswerSize, because only the server knows how the client will
// read the rest.
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
