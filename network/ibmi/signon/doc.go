// Package signon validates an IBM i password credential with the Sign-on
// host server (as-signon / as-signon-s) and reports the negotiated,
// non-secret system metadata.
//
// Validate runs one short-lived connection:
//
//	0x7003 exchange attributes  -> 0xF003 VRM, server level, password level, server seed
//	derive the password substitute locally from fresh seeds
//	0x7004 retrieve information -> 0xF004 return code, system CCSID
//	0x7006 end server (one-way), then close
//
// It authenticates once and never retries: a rejected credential produces
// exactly one 0x7004 request, because repeated failures can disable the
// profile (QMAXSIGN). The Sign-on seeds and substitute are never reused for
// the Remote Command connection, which runs its own seed exchange.
//
// Evidence: iSeriesNavigatorReverse docs/05-signon-and-authentication.md and
// docs/10-v5r4-compatibility-and-golden-vectors.md.
package signon
