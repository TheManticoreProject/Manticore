package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_ALIAS_ENUM_STRUCT is the server alias enumeration structure ([MS-SRVS]
// 2.2.4.106). Level selects the arm of the embedded SERVER_ALIAS_ENUM_UNION
// (which carries its own switch discriminant).
type SERVER_ALIAS_ENUM_STRUCT struct {
	Level           ndr.DWORD
	ServerAliasInfo SERVER_ALIAS_ENUM_UNION
}
