package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSAPR_TRUSTED_CONTROLLERS_INFO is an obsolete trusted-domain information class
// ([MS-LSAD] 2.2.7.5). Names is a [unique] pointer to a conformant array of
// RPC_UNICODE_STRING sized by Entries.
type LSAPR_TRUSTED_CONTROLLERS_INFO struct {
	Entries ndr.DWORD
	Names   []msdtyp.RPC_UNICODE_STRING `ndr:"unique,size_is=Entries"`
}
