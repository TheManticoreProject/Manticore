package data_structures

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// ULARGE_INTEGER
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/d37e0ce7-a358-4c07-a5c4-59c8b5da8b08
type ULARGE_INTEGER struct {
	// QuadPart: The 64-bit value.
	QuadPart data_types.UINT64
}

type PULARGE_INTEGER *ULARGE_INTEGER
