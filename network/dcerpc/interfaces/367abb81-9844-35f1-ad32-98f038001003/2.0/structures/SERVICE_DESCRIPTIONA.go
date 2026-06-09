package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVICE_DESCRIPTIONA ([MS-SCMR] 2.2.34) carries the ANSI service description as a
// [unique] pointer to a conformant-varying ASCII array.
type SERVICE_DESCRIPTIONA struct {
	LpDescription *ndr.STR `ndr:"unique"`
}
