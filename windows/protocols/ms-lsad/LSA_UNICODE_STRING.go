package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSA_UNICODE_STRING is the counted UTF-16 string used by the forest-trust structures
// ([MS-LSAD]); it is layout-identical to RPC_UNICODE_STRING ([MS-DTYP] 2.3.10), so it is a
// type alias rather than a redefinition.
type LSA_UNICODE_STRING = dtyp.RPC_UNICODE_STRING
