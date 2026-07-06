package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_GROUP_GENERAL_INFORMATION contains group fields ([MS-SAMR] 2.2.5.5). It holds the
// group's name, attributes, member count, and administrative comment.
type SAMPR_GROUP_GENERAL_INFORMATION struct {
	Name         msdtyp.RPC_UNICODE_STRING
	Attributes   ndr.DWORD
	MemberCount  ndr.DWORD
	AdminComment msdtyp.RPC_UNICODE_STRING
}
