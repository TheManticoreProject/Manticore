package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// DOMAIN_PASSWORD_INFORMATION contains a domain's password policy ([MS-SAMR] 2.2.4.5).
// MaxPasswordAge and MinPasswordAge are OLD_LARGE_INTEGER values defined by the base
// family.
type DOMAIN_PASSWORD_INFORMATION struct {
	MinPasswordLength     uint16
	PasswordHistoryLength uint16
	PasswordProperties    ndr.DWORD
	MaxPasswordAge        OLD_LARGE_INTEGER
	MinPasswordAge        OLD_LARGE_INTEGER
}
