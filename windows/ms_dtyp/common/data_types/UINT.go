package data_types

// A UINT is a 32-bit unsigned integer (range: 0 through 4294967295 decimal).
// Because a UINT is unsigned, its first bit (Most Significant Bit (MSB)) is not reserved for signing.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/52ddd4c3-55b9-4e03-8287-5392aac0627f
type UINT = uint32
