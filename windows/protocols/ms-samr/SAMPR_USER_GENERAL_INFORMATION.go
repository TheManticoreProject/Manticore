package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_GENERAL_INFORMATION holds general user attributes ([MS-SAMR]
// 2.2.6.7).
type SAMPR_USER_GENERAL_INFORMATION struct {
	UserName       msdtyp.RPC_UNICODE_STRING
	FullName       msdtyp.RPC_UNICODE_STRING
	PrimaryGroupId ndr.DWORD
	AdminComment   msdtyp.RPC_UNICODE_STRING
	UserComment    msdtyp.RPC_UNICODE_STRING
}
