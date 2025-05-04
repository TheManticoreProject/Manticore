package data_types

// A UINT32 is a 32-bit unsigned integer (range: 0 through 4294967295 decimal).
// Because a UINT32 is unsigned, its first bit (Most Significant Bit (MSB)) is not reserved for signing.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/362b8d72-2245-4c33-84a8-08c69f4c302f
type UINT32 = uint32
