package data_types

// An LPCSTR is a 32-bit pointer to a constant null-terminated string of 8-bit Windows (ANSI) characters.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/f8d4fe46-6be8-44c9-8823-615a21d17a61
type LPCSTR = *CHAR
