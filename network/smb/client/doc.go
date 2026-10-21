// Package client provides a single, version-agnostic SMB client.
//
// The caller constructs one [Client], supplies an ordered preference list of
// protocol versions, and dials. The client negotiates the dialect once and
// binds a version-specific backend (SMB1, SMB2, or — in future — SMB3) behind a
// common interface. Every subsequent call (TreeConnect, OpenFile, ReadFile, …)
// is routed to that backend, so the caller never has to know which dialect was
// selected.
//
// # Backends
//
// Each SMB engine (network/smb/smb_v10, network/smb/smb_v20, …) is wrapped by a
// thin adapter that satisfies the Backend interface. The dependency direction is
// one-way: this package imports the engines; the engines never import this
// package. Adapters live here.
//
// # Negotiation and preference
//
// Dialect selection is driven by a client-side preference list (Options.Preferred,
// highest priority first) and a policy:
//
//   - PolicyStrictOrder (default) tries the protocol families in the order the
//     list names them. It can force SMB1 on a server that also supports SMB3.
//   - PolicyHighestInSet performs a single multi-protocol negotiate over the
//     whole set and uses the server's highest-supported dialect within it.
//
// Under either policy only the listed dialects are offered: a dialect left out
// of Options.Preferred is never negotiated.
//
// PolicyStrictOrder makes one native negotiate per family, reconnecting between
// attempts and binding the first the server accepts. All the SMB2 dialects of
// the list share a single attempt, made at the position of the first of them,
// whose NEGOTIATE offers exactly those dialects; within that set the server
// selects its highest ([MS-SMB2] 3.3.5.4). So [SMB 2.1] yields SMB 2.1 even on
// an SMB 3.1.1 server, and [SMB 1.0, SMB 3.1.1] yields SMB1 on a server that
// speaks both.
//
// PolicyHighestInSet uses the SMB1 multi-protocol negotiate: one request offers
// the SMB1 and SMB2 dialect markers, and the reply is dispatched on its protocol
// marker (\xFFSMB for SMB1, \xFESMB for SMB2). "SMB 2.002" is offered when 2.0.2
// is listed, and the "SMB 2.???" wildcard when any later SMB2 dialect is; a
// wildcard reply is followed by a native SMB2 negotiate offering the listed SMB2
// dialects, which pins the exact revision.
//
// # Transports
//
// Options.Transport selects Direct TCP (the default, port 445) or NetBIOS over
// TCP (TransportNetBIOS, port 139), which performs the RFC 1002 session
// handshake before negotiating. SMB 3.1.1 is never offered over NetBIOS
// ([MS-SMB2] 3.2.5.2).
//
// Backends supported today: SMB 1.0 (network/smb/smb_v10) and SMB 2.0.2 through
// 3.1.1 (network/smb/smb_v20).
package client
