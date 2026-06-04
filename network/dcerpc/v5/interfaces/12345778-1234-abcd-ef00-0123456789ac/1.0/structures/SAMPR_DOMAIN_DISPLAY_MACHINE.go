package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_DOMAIN_DISPLAY_MACHINE holds a single machine entry returned by a display query
// ([MS-SAMR] 2.2.8.3).
type SAMPR_DOMAIN_DISPLAY_MACHINE struct {
	Index          ndr.DWORD
	Rid            ndr.DWORD
	AccountControl ndr.DWORD
	AccountName    dtyp.RPC_UNICODE_STRING
	AdminComment   dtyp.RPC_UNICODE_STRING
}
