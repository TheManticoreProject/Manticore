package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// USER_PRIMARY_GROUP_INFORMATION holds a user's primary group RID ([MS-SAMR]
// 2.2.6.9).
type USER_PRIMARY_GROUP_INFORMATION struct {
	PrimaryGroupId ndr.DWORD
}
