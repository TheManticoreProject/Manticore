package data_structures

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// LARGE_INTEGER
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/e904b1ba-f774-4203-ba1b-66485165ab1a
type LARGE_INTEGER struct {
	QuadPart data_types.ULONGLONG
}

type PLARGE_INTEGER *LARGE_INTEGER
