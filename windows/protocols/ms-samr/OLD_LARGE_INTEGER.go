package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// OLD_LARGE_INTEGER is a 64-bit value split into low and high parts
// ([MS-SAMR] 2.2.2.1).
type OLD_LARGE_INTEGER struct {
	LowPart  ndr.DWORD
	HighPart int32
}
