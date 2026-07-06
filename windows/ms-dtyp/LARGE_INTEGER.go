package msdtyp

// LARGE_INTEGER is the [MS-DTYP] 2.3.5 signed 64-bit integer. It is a named type rather
// than a bare int64 so declarations read like the IDL and it carries 8-octet NDR
// alignment ([C706] section 14.2.2).
//
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/e904b1ba-f774-4203-ba1b-66485165ab1a
type LARGE_INTEGER int64

type PLARGE_INTEGER *LARGE_INTEGER
