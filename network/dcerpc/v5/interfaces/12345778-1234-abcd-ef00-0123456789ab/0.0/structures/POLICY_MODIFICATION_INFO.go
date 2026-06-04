package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// POLICY_MODIFICATION_INFO is an obsolete policy information class ([MS-LSAD] 2.2.4.10).
type POLICY_MODIFICATION_INFO struct {
	ModifiedId           dtyp.LARGE_INTEGER
	DatabaseCreationTime dtyp.LARGE_INTEGER
}
