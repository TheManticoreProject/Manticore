package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_DOMAIN_GENERAL_INFORMATION2 extends SAMPR_DOMAIN_GENERAL_INFORMATION with
// account-lockout policy ([MS-SAMR] 2.2.4.10). LockoutDuration and
// LockoutObservationWindow are LARGE_INTEGER values from [MS-DTYP].
type SAMPR_DOMAIN_GENERAL_INFORMATION2 struct {
	I1                       SAMPR_DOMAIN_GENERAL_INFORMATION
	LockoutDuration          dtyp.LARGE_INTEGER
	LockoutObservationWindow dtyp.LARGE_INTEGER
	LockoutThreshold         uint16
}
