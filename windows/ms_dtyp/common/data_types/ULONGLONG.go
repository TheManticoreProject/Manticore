package data_types

// A ULONGLONG is a 64-bit unsigned integer (range: 0 through 18446744073709551615 decimal).
// Because a ULONGLONG is unsigned, its first bit (Most Significant Bit (MSB)) is not reserved for signing.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/c57d9fba-12ef-4853-b0d5-a6f472b50388
type ULONGLONG = uint64
