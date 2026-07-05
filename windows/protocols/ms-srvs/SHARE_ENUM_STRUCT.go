package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_ENUM_STRUCT contains a share enumeration level and the matching union of
// containers ([MS-SRVS] 2.2.4.39). The union arm is selected by Level via switch_is.
type SHARE_ENUM_STRUCT struct {
	Level     ndr.DWORD
	ShareInfo SHARE_ENUM_UNION
}
