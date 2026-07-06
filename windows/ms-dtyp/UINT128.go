package msdtyp

// UINT128
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/3b5c1a61-ece4-4dce-9c9f-c15388ba9032
type UINT128 struct {
	// Lower: The lower 64 bits of the 128-bit value.
	Lower UINT64
	// Upper: The upper 64 bits of the 128-bit value.
	Upper UINT64
}

type PUINT128 *UINT128
