package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_ALIAS_GENERAL_INFORMATION contains alias fields ([MS-SAMR] 2.2.6.2). It holds the
// alias's name, member count, and administrative comment.
type SAMPR_ALIAS_GENERAL_INFORMATION struct {
	Name         msdtyp.RPC_UNICODE_STRING
	MemberCount  ndr.DWORD
	AdminComment msdtyp.RPC_UNICODE_STRING
}
