package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_ALIAS_ENUM_UNION is the [switch_is(Level)] union declared inline in
// SERVER_ALIAS_ENUM_STRUCT ([MS-SRVS] 2.2.4.106), modeled here as a named type.
// Tag carries the discriminant (the level) inline, followed by the selected
// arm. The case-0 arm is a [unique] pointer to a SERVER_ALIAS_INFO_0_CONTAINER.
type SERVER_ALIAS_ENUM_UNION struct {
	Tag    ndr.DWORD                      `ndr:"switch"`
	Level0 *SERVER_ALIAS_INFO_0_CONTAINER `ndr:"case=0,unique"`
}
