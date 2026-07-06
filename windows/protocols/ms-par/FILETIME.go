package mspar

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// FILETIME is the [MS-DTYP] 2.3.3 64-bit timestamp. It is the shared msdtyp type, not a
// per-protocol redefinition: MS-PAR references it but does not define it in its own IDL.
type FILETIME = msdtyp.FILETIME
