package msswn

// PCONTEXT_HANDLE_SHARED is the shared RPC context handle passed to WitnessrAsyncNotify
// ([MS-SWN] 2.2.1.3). The IDL types it as [context_handle] PCONTEXT_HANDLE, allowing the
// same registration handle to be used concurrently by the long-running AsyncNotify call
// while other calls run; on the wire it is the same 20-octet context handle as
// PCONTEXT_HANDLE, so it is modeled as an alias.
type PCONTEXT_HANDLE_SHARED = PCONTEXT_HANDLE
