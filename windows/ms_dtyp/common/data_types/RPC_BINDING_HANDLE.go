package data_types

// An RPC_BINDING_HANDLE is an untyped 32-bit pointer containing information that the RPC run-time library uses to access binding information.
// It is directly equivalent to the type rpc_binding_handle_t described in [C706] section 3.1.4.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/d0dffa33-812f-4214-a987-4ee149328eec
type RPC_BINDING_HANDLE = uintptr
