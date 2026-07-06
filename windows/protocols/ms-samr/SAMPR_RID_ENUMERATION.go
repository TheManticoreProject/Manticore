package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_RID_ENUMERATION pairs a relative ID with the principal's name
// ([MS-SAMR] 2.2.3.4).
type SAMPR_RID_ENUMERATION struct {
	RelativeId ndr.DWORD
	Name       msdtyp.RPC_UNICODE_STRING
}
