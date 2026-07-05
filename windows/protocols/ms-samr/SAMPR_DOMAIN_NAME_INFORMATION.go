package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_DOMAIN_NAME_INFORMATION contains the name of a domain ([MS-SAMR] 2.2.4.12).
type SAMPR_DOMAIN_NAME_INFORMATION struct {
	DomainName dtyp.RPC_UNICODE_STRING
}
