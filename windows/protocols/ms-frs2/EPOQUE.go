package msfrs2

import "github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"

// EPOQUE is a SYSTEMTIME used to version epoque vectors ([MS-FRS2] 2.2.1.4.7). It is an
// alias of dtyp.SYSTEMTIME so it marshals identically on the wire.
type EPOQUE = dtyp.SYSTEMTIME
