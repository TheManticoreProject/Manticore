package msswn

// PPCONTEXT_HANDLE is the [out] / [in,out] context-handle parameter type of
// WitnessrRegister, WitnessrRegisterEx, and WitnessrUnRegisterEx ([MS-SWN] 2.2.1.2). The
// IDL types it as [ref] PCONTEXT_HANDLE * — a ref pointer to a context handle. A ref
// pointer to a context handle emits no referent id, and the context handle itself is
// transmitted inline as 20 octets, so on the wire it is identical to PCONTEXT_HANDLE and
// is modeled as an alias.
type PPCONTEXT_HANDLE = PCONTEXT_HANDLE
