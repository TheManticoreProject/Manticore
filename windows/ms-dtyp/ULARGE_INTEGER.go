package msdtyp

// ULARGE_INTEGER is the [MS-DTYP] 2.3.13 unsigned 64-bit integer. It is a named type
// rather than a bare uint64 so declarations read like the IDL and it carries 8-octet NDR
// alignment ([C706] section 14.2.2).
//
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/d7e6e5a5-6c77-4ae6-9bd5-3892b3c4641e
type ULARGE_INTEGER uint64

type PULARGE_INTEGER *ULARGE_INTEGER
