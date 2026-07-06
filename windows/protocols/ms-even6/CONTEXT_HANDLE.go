package mseven6

// CONTEXT_HANDLE is the generic 20-octet RPC context handle ([MS-RPCE] 2.3.2.2) used by
// EvtRpcClose. The IDL types it as an [in, out] context_handle void**, so it accepts any
// of the interface's typed handles (log query, log handle, operation control,
// subscription, …); callers convert their typed [20]byte handle to this type. It is
// transmitted inline (no referent id), matching every other context handle in this
// interface.
type CONTEXT_HANDLE [20]byte
