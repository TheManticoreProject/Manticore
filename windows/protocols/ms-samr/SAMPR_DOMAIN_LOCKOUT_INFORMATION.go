package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_DOMAIN_LOCKOUT_INFORMATION contains a domain's account-lockout policy
// ([MS-SAMR] 2.2.4.14). LockoutDuration and LockoutObservationWindow are LARGE_INTEGER
// values from [MS-DTYP].
type SAMPR_DOMAIN_LOCKOUT_INFORMATION struct {
	LockoutDuration          dtyp.LARGE_INTEGER
	LockoutObservationWindow dtyp.LARGE_INTEGER
	LockoutThreshold         uint16
}
