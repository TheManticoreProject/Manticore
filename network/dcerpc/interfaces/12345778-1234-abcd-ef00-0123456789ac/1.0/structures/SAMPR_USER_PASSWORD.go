package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_USER_PASSWORD holds a cleartext password and its length ([MS-SAMR]
// 2.2.6.20). Buffer is a fixed 256-WCHAR array; Length is the number of bytes
// of password data in Buffer (counted from the end).
type SAMPR_USER_PASSWORD struct {
	Buffer [256]uint16
	Length ndr.DWORD
}
