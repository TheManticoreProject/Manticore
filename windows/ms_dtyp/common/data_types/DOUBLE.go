package data_types

// A DOUBLE is an 8-byte, double-precision, floating-point number that represents a double-precision, 64-bit [IEEE754]
// value with the approximate range: +/-5.0 x 10-324 through +/-1.7 x 10308.
// The DOUBLE type can also represent not a number (NAN); positive and negative infinity; or positive and negative 0.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/40beeef3-303a-40de-895c-11379fc42c15
type DOUBLE = float64
