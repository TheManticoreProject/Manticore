package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_AUTH_INFORMATION carries a single trust authentication entry ([MS-LSAD]
// 2.2.7.17). AuthInfo is a [unique] pointer to a conformant byte array sized by
// AuthInfoLength.
type LSAPR_AUTH_INFORMATION struct {
	LastUpdateTime dtyp.LARGE_INTEGER
	AuthType       ndr.DWORD
	AuthInfoLength ndr.DWORD
	AuthInfo       []byte `ndr:"unique,size_is=AuthInfoLength"`
}
