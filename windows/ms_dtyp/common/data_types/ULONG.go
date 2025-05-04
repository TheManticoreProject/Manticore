package data_types

// A ULONG is a 32-bit unsigned integer (range: 0 through 4294967295 decimal). Because a ULONG is unsigned, its first bit (Most Significant Bit (MSB)) is not reserved for signing.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/32862b84-f6e6-40f9-85ca-c4faf985b822
type ULONG = uint32
