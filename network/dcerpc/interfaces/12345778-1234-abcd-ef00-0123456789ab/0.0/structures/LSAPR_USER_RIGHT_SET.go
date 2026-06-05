package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_USER_RIGHT_SET is a set of user-right (privilege/system-access) names ([MS-LSAD]
// 2.2.5.5). UserRights is a [unique] pointer to a conformant array of RPC_UNICODE_STRING
// (the IDL element type PRPC_UNICODE_STRING resolves to the struct, transmitted inline)
// sized by Entries.
type LSAPR_USER_RIGHT_SET struct {
	Entries    ndr.DWORD
	UserRights []dtyp.RPC_UNICODE_STRING `ndr:"unique,size_is=Entries"`
}
