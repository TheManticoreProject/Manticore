package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// CONNECT_ENUM_STRUCT is the connection enumeration structure ([MS-SRVS] 2.2.4.6).
// Level selects the arm of the embedded CONNECT_ENUM_UNION (which carries its own
// switch discriminant).
type CONNECT_ENUM_STRUCT struct {
	Level       ndr.DWORD
	ConnectInfo CONNECT_ENUM_UNION
}
