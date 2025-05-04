package data_types

// NTSTATUS is a standard 32-bit datatype for system-supplied status code values.
// NTSTATUS values are used to communicate system information. They are of four types:
// success values, information values, warnings, and error values, as specified in [MS-ERREF].
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/c8b512d5-70b1-4028-95f1-ec92d35cb51e
type NTSTATUS = LONG
