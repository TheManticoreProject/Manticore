package data_types

// A UINT64 is a 64-bit unsigned integer (range: 0 through 18446744073709551615 decimal).
// Because a UINT64 is unsigned, its first bit (Most Significant Bit (MSB)) is not reserved for signing.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/a7b7720f-87eb-4add-9bcb-c6ff652778ae
type UINT64 = uint64
