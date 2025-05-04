package data_types

// A WORD is a 16-bit unsigned integer (range: 0 through 65535 decimal).
// Because a WORD is unsigned, its first bit (Most Significant Bit (MSB)) is not reserved for signing.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/f8573df3-a44a-4a50-b070-ac4c3aa78e3c
type WORD = uint16
