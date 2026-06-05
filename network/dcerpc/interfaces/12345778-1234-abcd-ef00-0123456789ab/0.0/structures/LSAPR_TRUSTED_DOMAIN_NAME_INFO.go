package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_TRUSTED_DOMAIN_NAME_INFO contains the name of a trusted domain ([MS-LSAD]
// 2.2.7.4).
type LSAPR_TRUSTED_DOMAIN_NAME_INFO struct {
	Name dtyp.RPC_UNICODE_STRING
}
