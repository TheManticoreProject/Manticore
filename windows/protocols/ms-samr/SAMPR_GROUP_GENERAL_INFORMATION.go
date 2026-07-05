package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_GROUP_GENERAL_INFORMATION contains group fields ([MS-SAMR] 2.2.5.5). It holds the
// group's name, attributes, member count, and administrative comment.
type SAMPR_GROUP_GENERAL_INFORMATION struct {
	Name         dtyp.RPC_UNICODE_STRING
	Attributes   ndr.DWORD
	MemberCount  ndr.DWORD
	AdminComment dtyp.RPC_UNICODE_STRING
}
