package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// GROUP_MEMBERSHIP pairs a group relative ID with its attributes ([MS-SAMR]
// 2.2.3.6).
type GROUP_MEMBERSHIP struct {
	RelativeId ndr.DWORD
	Attributes ndr.DWORD
}
