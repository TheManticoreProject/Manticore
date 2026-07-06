package msnspi

import msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"

// FILETIME is the [MS-DTYP] FILETIME (2.3.3): a 64-bit count of 100-nanosecond
// intervals since January 1, 1601 (UTC), split into two 32-bit words. The NSPI IDL
// imports it from ms-msdtyp.idl rather than defining it, so it is aliased to msdtyp.FILETIME.
type FILETIME = msdtyp.FILETIME
