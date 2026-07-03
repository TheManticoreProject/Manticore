package mspan

// PNOTIFYOBJECT is a notification-channel RPC context handle ([MS-PAN] 3.1.1): an
// [context_handle] void* carried on the wire as a 20-octet handle ([MS-RPCE] 2.3.2.2).
type PNOTIFYOBJECT [20]byte
