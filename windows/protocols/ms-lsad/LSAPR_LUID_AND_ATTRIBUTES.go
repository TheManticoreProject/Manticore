package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_LUID_AND_ATTRIBUTES is a privilege LUID together with its attribute flags
// ([MS-LSAD] 2.2.5.1).
type LSAPR_LUID_AND_ATTRIBUTES struct {
	Luid       dtyp.LUID
	Attributes ndr.DWORD
}
