package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// USER_CONTROL_INFORMATION holds a user's account control bits ([MS-SAMR]
// 2.2.6.14).
type USER_CONTROL_INFORMATION struct {
	UserAccountControl ndr.DWORD
}
