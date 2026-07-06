package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_DOMAIN_LOCKOUT_INFORMATION contains a domain's account-lockout policy
// ([MS-SAMR] 2.2.4.14). LockoutDuration and LockoutObservationWindow are LARGE_INTEGER
// values from [MS-DTYP].
type SAMPR_DOMAIN_LOCKOUT_INFORMATION struct {
	LockoutDuration          msdtyp.LARGE_INTEGER
	LockoutObservationWindow msdtyp.LARGE_INTEGER
	LockoutThreshold         uint16
}
