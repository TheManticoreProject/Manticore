package msfrs2

import msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"

// EPOQUE is a SYSTEMTIME used to version epoque vectors ([MS-FRS2] 2.2.1.4.7). It is an
// alias of msdtyp.SYSTEMTIME so it marshals identically on the wire.
type EPOQUE = msdtyp.SYSTEMTIME
