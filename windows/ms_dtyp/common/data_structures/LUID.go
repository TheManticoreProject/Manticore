package data_structures

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// The LUID structure is 64-bit value guaranteed to be unique only on the system on which it was generated.
// The uniqueness of a locally unique identifier (LUID) is guaranteed only until the system is restarted.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/48cbee2a-0790-45f2-8269-931d7083b2c3
type LUID struct {
	// LowPart: The low-order bits of the structure.
	LowPart data_types.DWORD
	// HighPart: The high-order bits of the structure.
	HighPart data_types.LONG
}

type PLUID *LUID
