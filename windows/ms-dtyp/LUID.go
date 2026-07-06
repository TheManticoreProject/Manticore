package msdtyp

// The LUID structure is 64-bit value guaranteed to be unique only on the system on which it was generated.
// The uniqueness of a locally unique identifier (LUID) is guaranteed only until the system is restarted.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/48cbee2a-0790-45f2-8269-931d7083b2c3
type LUID struct {
	// LowPart: The low-order bits of the structure.
	LowPart DWORD
	// HighPart: The high-order bits of the structure.
	HighPart LONG
}

type PLUID *LUID

// Uint64 returns the LUID as a single 64-bit value (HighPart in the upper 32 bits).
func (l LUID) Uint64() uint64 {
	return uint64(uint32(l.HighPart))<<32 | uint64(l.LowPart)
}

// LUIDFromUint64 splits a 64-bit value into a LUID.
func LUIDFromUint64(v uint64) LUID {
	return LUID{LowPart: uint32(v), HighPart: int32(v >> 32)}
}
