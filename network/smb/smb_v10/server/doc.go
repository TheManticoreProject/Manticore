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
//   - SMB_COM_ECHO.
//
// Not yet implemented: every other command. A request carrying one is answered
// with STATUS_NOT_IMPLEMENTED, which means in particular that no client can yet
// negotiate a dialect, authenticate, or reach a share. Negotiation and
// authentication arrive with the phases that add them.
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
