package data_structures

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// A language code identifier structure is stored as a DWORD. The lower word contains the language identifier,
// and the upper word contains both the sorting identifier (ID) and a reserved value. For additional details about
// the structure and possible values, see [MS-LCID].
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/e8e4255f-5b6d-472b-8a98-ae3950bfdb9a
type LCID = data_types.DWORD
