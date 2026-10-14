// Package hostserver implements the framing common to IBM i Client Access
// host servers: the 20-byte big-endian header, bounded frames, LL/CP
// variable fields, correlation IDs, and a frame transport over TCP or
// verified TLS.
//
// It knows nothing about authentication or any request's meaning. Sign-on
// and Remote Command build on it.
//
// Every peer-controlled length is checked against an explicit cap before
// allocation. A Conn that fails mid-frame, or receives a malformed or
// unexpected frame, is poisoned and closed: the byte stream cannot be
// resynchronized, so it is never reused. Tracing reports metadata only:
// never frame bytes.
package hostserver
