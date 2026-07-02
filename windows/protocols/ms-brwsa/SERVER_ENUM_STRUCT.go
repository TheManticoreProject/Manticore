package msbrwsa

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_ENUM_STRUCT holds the level and the level-tagged container returned by
// I_BrowserrQueryOtherDomains ([MS-BRWSA] 2.2.3.2). Level MUST be 100. The switch_is(Level)
// union's discriminant is transmitted inline as SERVER_ENUM_UNION.Tag; set both Level and
// ServerInfo.Tag to 100 before marshalling a request.
type SERVER_ENUM_STRUCT struct {
	Level      ndr.DWORD
	ServerInfo SERVER_ENUM_UNION
}

// SERVER_ENUM_UNION is the [switch_is(Level)] union of server-enumeration containers
// ([MS-BRWSA] 2.2.3.2). Tag carries the discriminant (the level) inline, followed by the
// selected arm. The case-100 arm is a [unique] pointer to its container
// (LPSERVER_INFO_100_CONTAINER); the [default] arm carries nothing.
type SERVER_ENUM_UNION struct {
	Tag      ndr.DWORD                  `ndr:"switch"`
	Level100 *SERVER_INFO_100_CONTAINER `ndr:"case=100,unique"`
}
