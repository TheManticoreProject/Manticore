package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_DOMAIN_OEM_INFORMATION contains the OEM-defined comment string for a domain
// ([MS-SAMR] 2.2.4.11).
type SAMPR_DOMAIN_OEM_INFORMATION struct {
	OemInformation dtyp.RPC_UNICODE_STRING
}
