package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_ALIAS_INFO is the [switch_type(unsigned long)] union used by the
// server alias Get and Set Info methods ([MS-SRVS] 2.2.3.8). Tag carries the
// discriminant (the level) inline, followed by the selected arm. The case-0
// arm is a [unique] pointer to a SERVER_ALIAS_INFO_0.
type SERVER_ALIAS_INFO struct {
	Tag              ndr.DWORD            `ndr:"switch"`
	ServerAliasInfo0 *SERVER_ALIAS_INFO_0 `ndr:"case=0,unique"`
}
