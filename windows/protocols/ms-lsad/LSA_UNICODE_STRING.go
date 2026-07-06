package mslsad

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSA_UNICODE_STRING is the counted UTF-16 string used by the forest-trust structures
// ([MS-LSAD]); it is layout-identical to RPC_UNICODE_STRING ([MS-DTYP] 2.3.10), so it is a
// type alias rather than a redefinition.
type LSA_UNICODE_STRING = msdtyp.RPC_UNICODE_STRING
