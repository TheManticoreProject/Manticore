package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_USER_PASSWORD_NEW holds a cleartext password, its length, and a salt
// ([MS-SAMR] 2.2.6.23). Buffer is a fixed 256-WCHAR array; Length is the number
// of bytes of password data; ClearSalt is a 16-byte salt.
type SAMPR_USER_PASSWORD_NEW struct {
	Buffer    [256]uint16
	Length    ndr.DWORD
	ClearSalt [16]byte
}
