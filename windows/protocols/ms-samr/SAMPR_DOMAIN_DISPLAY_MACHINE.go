package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_DOMAIN_DISPLAY_MACHINE holds a single machine entry returned by a display query
// ([MS-SAMR] 2.2.8.3).
type SAMPR_DOMAIN_DISPLAY_MACHINE struct {
	Index          ndr.DWORD
	Rid            ndr.DWORD
	AccountControl ndr.DWORD
	AccountName    msdtyp.RPC_UNICODE_STRING
	AdminComment   msdtyp.RPC_UNICODE_STRING
}
