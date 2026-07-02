package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_PRIVILEGE_ENUM_BUFFER is the result buffer of a privilege enumeration
// ([MS-LSAD] 2.2.8.2). Privileges is a [unique] pointer to a conformant array of
// LSAPR_POLICY_PRIVILEGE_DEF sized by Entries.
type LSAPR_PRIVILEGE_ENUM_BUFFER struct {
	Entries    ndr.DWORD
	Privileges []LSAPR_POLICY_PRIVILEGE_DEF `ndr:"unique,size_is=Entries"`
}
