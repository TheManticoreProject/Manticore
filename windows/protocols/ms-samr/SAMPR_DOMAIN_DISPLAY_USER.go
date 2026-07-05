package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_DOMAIN_DISPLAY_USER holds a single user entry returned by a display query
// ([MS-SAMR] 2.2.8.2).
type SAMPR_DOMAIN_DISPLAY_USER struct {
	Index          ndr.DWORD
	Rid            ndr.DWORD
	AccountControl ndr.DWORD
	AccountName    dtyp.RPC_UNICODE_STRING
	AdminComment   dtyp.RPC_UNICODE_STRING
	FullName       dtyp.RPC_UNICODE_STRING
}
