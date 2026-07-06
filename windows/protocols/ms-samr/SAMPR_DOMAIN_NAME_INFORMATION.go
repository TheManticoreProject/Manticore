package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_DOMAIN_NAME_INFORMATION contains the name of a domain ([MS-SAMR] 2.2.4.12).
type SAMPR_DOMAIN_NAME_INFORMATION struct {
	DomainName msdtyp.RPC_UNICODE_STRING
}
