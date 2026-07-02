package structures

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// FILETIME is the [MS-DTYP] 2.3.3 64-bit value (100-ns intervals since 1601) as two
// 32-bit halves. Referenced by the driver info structures but not defined in the MS-RPRN IDL.
type FILETIME struct {
	DwLowDateTime  ndr.DWORD
	DwHighDateTime ndr.DWORD
}
