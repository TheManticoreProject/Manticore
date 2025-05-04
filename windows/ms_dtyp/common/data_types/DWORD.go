package data_types

// A DWORD is a 32-bit unsigned integer (range: 0 through 4294967295 decimal). Because a DWORD is unsigned,
// its first bit (Most Significant Bit (MSB)) is not reserved for signing.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/262627d8-3418-4627-9218-4ffe110850b2
type DWORD = uint32
