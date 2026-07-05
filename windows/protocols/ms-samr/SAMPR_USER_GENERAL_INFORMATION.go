package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_USER_GENERAL_INFORMATION holds general user attributes ([MS-SAMR]
// 2.2.6.7).
type SAMPR_USER_GENERAL_INFORMATION struct {
	UserName       dtyp.RPC_UNICODE_STRING
	FullName       dtyp.RPC_UNICODE_STRING
	PrimaryGroupId ndr.DWORD
	AdminComment   dtyp.RPC_UNICODE_STRING
	UserComment    dtyp.RPC_UNICODE_STRING
}
