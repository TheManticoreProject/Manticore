package data_types

// An LPCWSTR is a 32-bit pointer to a constant string of 16-bit Unicode characters, which MAY be null-terminated.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/76f10dd8-699d-45e6-a53c-5aefc586da20
type LPCWSTR = *WCHAR
