package data_structures

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// The MULTI_SZ structure defines an implementation-specific<4> type that contains a sequence of null-terminated strings,
// terminated by an empty string (\0) so that the last two characters are both null terminators.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/fd7b2d81-b1d7-414f-a3df-c66fabc578db
type MULTI_SZ struct {
	// Value: A data buffer, which is a string literal containing multiple null-terminated strings serially.
	Value data_types.WCHAR
	// NChar: The length, in characters, including the two terminating nulls.
	NChar data_types.DWORD
}

type PMULTI_SZ *MULTI_SZ
