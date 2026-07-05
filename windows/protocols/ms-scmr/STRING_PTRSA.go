package msscmr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// STRING_PTRSA ([MS-SCMR] 2.2.21) wraps a single [unique] ANSI string argument for
// RStartServiceA. StringPtr is a [unique] pointer to a conformant-varying ASCII array.
type STRING_PTRSA struct {
	StringPtr *ndr.STR `ndr:"unique"`
}
