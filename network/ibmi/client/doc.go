// Package client is the public IBM i Client Access connection facade.
//
// New validates a Config without I/O. Connect validates the credential once
// with the Sign-on server, then starts an authenticated Remote Command
// server job with its own fresh seeds; the credential is not retained after
// Connect returns. SessionInfo returns an immutable, non-secret snapshot.
// Close ends the server job and closes every socket; it is idempotent.
//
// Lifecycle:
//
//	new -> connecting -> ready -> closing -> closed
//	                       \-> poisoned (framing became ambiguous) -> closed
//	a failed Connect     -> closed (create a new Client to retry)
//
// All operations on the session hold one lifecycle lock, so Close waits for
// an in-flight operation and a compound operation cannot be interleaved.
//
// Transport security is explicit: Config.Transport must be TLSRequired
// (verified TLS, never downgraded) or CleartextExplicit (isolated labs
// only). Authentication failures are returned, never retried, because
// repeated failures can disable the profile.
package client
