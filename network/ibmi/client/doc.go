// Package client is the public IBM i Client Access connection facade.
//
// New validates a Config without I/O. Connect validates the credential (a
// password, or a multiple-use profile token) once with the Sign-on server,
// then starts an authenticated Remote Command server job, deriving a fresh
// password substitute for it; the credential is not retained after Connect
// returns. GenerateProfileToken issues a profile token from a password. SessionInfo returns an immutable, non-secret snapshot.
// Close ends the server job and closes every socket; it is idempotent.
//
// Lifecycle:
//
//	new -> connecting -> ready -> closing -> closed
//	                       \-> poisoned (framing became ambiguous) -> closed
//	a failed Connect     -> closed (create a new Client to retry)
//
// Client implements rpc.Caller (Call) and rpc.ExclusiveCaller
// (ExecuteExclusive), so typed IBM i API packages can depend on those
// interfaces instead of this facade. RunCommand runs raw CL and is a
// privileged operation.
//
// All operations on the session hold one lifecycle lock, so Close waits for
// an in-flight operation and a compound operation cannot be interleaved.
//
// Transport security is explicit: Config.Transport must be TLSRequired
// (verified TLS, never downgraded) or CleartextExplicit (isolated labs
// only). Authentication failures are returned, never retried, because
// repeated failures can disable the profile.
package client
