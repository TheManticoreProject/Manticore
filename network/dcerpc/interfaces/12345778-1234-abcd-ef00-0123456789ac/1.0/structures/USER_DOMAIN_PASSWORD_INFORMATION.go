package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// USER_DOMAIN_PASSWORD_INFORMATION carries domain password policy constraints
// ([MS-SAMR] 2.2.3.14).
type USER_DOMAIN_PASSWORD_INFORMATION struct {
	MinPasswordLength  uint16
	PasswordProperties ndr.DWORD
}
