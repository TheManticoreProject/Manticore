package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_PRIVILEGE_SET is a set of privileges held by an account ([MS-LSAD] 2.2.5.2).
// Privilege is an inline conformant array of LSAPR_LUID_AND_ATTRIBUTES (the IDL declares
// it as an embedded array, not a pointer) sized by PrivilegeCount, so NDR hoists its
// maximum_count to the front of the structure.
type LSAPR_PRIVILEGE_SET struct {
	PrivilegeCount ndr.DWORD
	Control        ndr.DWORD
	Privilege      []LSAPR_LUID_AND_ATTRIBUTES `ndr:"conformant,size_is=PrivilegeCount"`
}
