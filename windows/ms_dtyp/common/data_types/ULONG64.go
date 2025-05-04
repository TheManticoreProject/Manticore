package data_types

// A ULONG64 is a 64-bit unsigned integer (range: 0 through 18446744073709551615 decimal).
// Because a ULONG64 is unsigned, its first bit (Most Significant Bit (MSB)) is not reserved for signing.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/2dc4c492-95db-4fa6-ae2b-8546b13c9141
type ULONG64 = uint64
