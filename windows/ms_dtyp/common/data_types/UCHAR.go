package data_types

// A UCHAR is an 8-bit integer with the range: 0 through 255 decimal. Because a UCHAR is unsigned, its first bit (Most Significant Bit (MSB)) is not reserved for signing.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/050baef1-f978-4851-a3c7-ad701a90e54a
type UCHAR = uint8
