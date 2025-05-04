package data_types

// The PCONTEXT_HANDLE type keeps state information associated with a given client on a server. The state information is called the server's context. Clients can obtain a context handle to identify the server's context for their individual RPC sessions.
// A context handle must be of the void * type, or a type that resolves to void *. The server program casts it to the required type.
// The IDL attribute [context_handle], as specified in [C706], is used to declare PCONTEXT_HANDLE.
// An interface that uses a context handle must have a binding handle for the initial binding, which has to take place before the server can return a context handle. The handle_t type is one of the predefined types of the interface definition language (IDL), which is used to create a binding handle.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/b8d08028-e73f-4bb8-b3bf-bc8ba4734c4c
type PCONTEXT_HANDLE = uintptr
