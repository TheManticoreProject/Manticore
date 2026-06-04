package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_POLICY_PRIVILEGE_DEF defines a privilege by name and its local LUID value
// ([MS-LSAD] 2.2.8.1).
type LSAPR_POLICY_PRIVILEGE_DEF struct {
	Name       dtyp.RPC_UNICODE_STRING
	LocalValue dtyp.LUID
}
