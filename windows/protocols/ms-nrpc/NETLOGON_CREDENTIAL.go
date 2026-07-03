// Package msnrpc holds the NDR wire structures of the Netlogon Remote Protocol
// ([MS-NRPC], interface 12345678-1234-abcd-ef00-01234567cffb version 1.0).
package msnrpc

// NETLOGON_CREDENTIAL ([MS-NRPC] 2.2.1.3.4) is an opaque 8-octet challenge or credential.
// The IDL declares it as a struct with a single CHAR data[8] field; on the wire that is a
// fixed array of 8 bytes with no NDR conformance framing, so a Go [8]byte marshals it
// exactly while staying ergonomic for the credential cryptography (indexable/sliceable).
type NETLOGON_CREDENTIAL [8]byte
