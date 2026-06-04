package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TRUSTED_POSIX_OFFSET_INFO communicates the POSIX offset of a trusted domain ([MS-LSAD]
// 2.2.7.6).
type TRUSTED_POSIX_OFFSET_INFO struct {
	Offset ndr.DWORD
}
