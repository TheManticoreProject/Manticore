package msnrpc

// NETLOGON_AUTHENTICATOR ([MS-NRPC] 2.2.1.3.5) carries a client/server credential and a
// timestamp. The 8-byte credential needs no alignment; the ULONG timestamp is 4-aligned,
// landing at offset 8 (already aligned), for a 12-octet structure.
type NETLOGON_AUTHENTICATOR struct {
	Credential NETLOGON_CREDENTIAL
	Timestamp  uint32
}
