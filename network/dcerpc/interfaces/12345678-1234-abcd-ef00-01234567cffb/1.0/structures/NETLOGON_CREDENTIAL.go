// Package structures holds the NDR wire types of the Netlogon interface
// (12345678-1234-abcd-ef00-01234567cffb, [MS-NRPC]).
package structures

// NETLOGON_CREDENTIAL ([MS-NRPC] 2.2.1.3.4) is an opaque 8-octet challenge or credential.
// On the wire it is a fixed array of 8 bytes with no NDR conformance framing, so a Go
// [8]byte marshals it exactly.
type NETLOGON_CREDENTIAL [8]byte
