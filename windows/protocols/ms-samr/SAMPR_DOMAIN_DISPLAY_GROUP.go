package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_DOMAIN_DISPLAY_GROUP holds a single group entry returned by a display query
// ([MS-SAMR] 2.2.8.4).
type SAMPR_DOMAIN_DISPLAY_GROUP struct {
	Index        ndr.DWORD
	Rid          ndr.DWORD
	Attributes   ndr.DWORD
	AccountName  msdtyp.RPC_UNICODE_STRING
	AdminComment msdtyp.RPC_UNICODE_STRING
}
