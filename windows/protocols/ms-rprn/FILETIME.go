package msrprn

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// FILETIME is the [MS-DTYP] 2.3.3 64-bit timestamp. It is the shared msdtyp type, not a
// per-protocol redefinition: MS-RPRN references it (in the driver info structures) but
// does not define it in its own IDL.
type FILETIME = msdtyp.FILETIME
