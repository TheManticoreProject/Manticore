package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_ENUM_STRUCT is the session enumeration structure ([MS-SRVS] 2.2.4.23).
// Level selects the arm of the embedded SESSION_ENUM_UNION (which carries its own
// switch discriminant).
type SESSION_ENUM_STRUCT struct {
	Level       ndr.DWORD
	SessionInfo SESSION_ENUM_UNION
}
