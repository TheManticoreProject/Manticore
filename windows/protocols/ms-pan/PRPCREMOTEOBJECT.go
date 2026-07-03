package mspan

// PRPCREMOTEOBJECT is the remote-object RPC context handle ([MS-PAN] 2.2.4): an
// [context_handle] void* carried on the wire as a 20-octet handle ([MS-RPCE] 2.3.2.2).
// It is created by IRPCRemoteObject_Create and consumed by the IRPCAsyncNotify methods.
type PRPCREMOTEOBJECT [20]byte
