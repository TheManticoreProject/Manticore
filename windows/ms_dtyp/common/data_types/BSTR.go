package data_types

// A BSTR is a pointer to a null-terminated character string in which the string length is stored with the string.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/692a42a9-06ce-4394-b9bc-5d2a50440168
type BSTR = []WCHAR
