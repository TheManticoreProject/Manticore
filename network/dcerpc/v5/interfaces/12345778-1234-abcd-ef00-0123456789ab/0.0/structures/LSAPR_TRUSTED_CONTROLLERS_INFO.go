package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRUSTED_CONTROLLERS_INFO is an obsolete trusted-domain information class
// ([MS-LSAD] 2.2.7.5). Names is a [unique] pointer to a conformant array of
// RPC_UNICODE_STRING sized by Entries.
type LSAPR_TRUSTED_CONTROLLERS_INFO struct {
	Entries ndr.DWORD
	Names   []dtyp.RPC_UNICODE_STRING `ndr:"unique,size_is=Entries"`
}
