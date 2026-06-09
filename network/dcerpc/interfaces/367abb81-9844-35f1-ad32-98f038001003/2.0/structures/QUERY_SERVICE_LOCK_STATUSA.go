package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// QUERY_SERVICE_LOCK_STATUSA ([MS-SCMR] 2.2.8) is the ANSI lock-status record. LpLockOwner
// is a [unique] pointer to a conformant-varying ASCII array.
type QUERY_SERVICE_LOCK_STATUSA struct {
	FIsLocked      ndr.DWORD
	LpLockOwner    *ndr.STR `ndr:"unique"`
	DwLockDuration ndr.DWORD
}
