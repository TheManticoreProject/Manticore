package mslsad

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSAPR_TRUSTED_DOMAIN_NAME_INFO contains the name of a trusted domain ([MS-LSAD]
// 2.2.7.4).
type LSAPR_TRUSTED_DOMAIN_NAME_INFO struct {
	Name msdtyp.RPC_UNICODE_STRING
}
