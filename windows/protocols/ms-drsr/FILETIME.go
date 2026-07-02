package msdrsr

import "github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"

// FILETIME is the [MS-DTYP] FILETIME (2.3.3): a 64-bit count of 100-nanosecond
// intervals since January 1, 1601 (UTC), split into two 32-bit words. The IDL imports
// it from ms-dtyp.idl rather than defining it, so it is aliased to dtyp.FILETIME.
type FILETIME = dtyp.FILETIME
