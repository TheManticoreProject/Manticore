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
//   - PolicyStrictOrder (default) tries the preferred versions in order and
//     selects the first one the server accepts. This honors the caller's order
//     exactly — it can force SMB1 on a server that also supports SMB3.
//   - PolicyHighestInSet performs a single multi-protocol negotiate over the
//     whole set and uses the server's highest-supported dialect within it.
//
// Discovery uses the SMB1 multi-protocol negotiate: one request offers the SMB1
// and SMB2 dialect markers, and the reply is dispatched on its protocol marker
// (\xFFSMB for SMB1, \xFESMB for SMB2); the SMB2 wildcard dialect triggers a
// second, native SMB2 negotiate to pin the exact 2.1/3.x revision.
//
// This package is built in phases; see .private/smb_common/docs for the plan.
package client
