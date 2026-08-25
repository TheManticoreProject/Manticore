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
//
// Not yet implemented: every other command, answered with
// STATUS_NOT_IMPLEMENTED. A client can now authenticate, but there is still
// nothing to reach: no share can be connected to and no file opened, because
// tree connect and the file commands arrive with a later phase.
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
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
package server
