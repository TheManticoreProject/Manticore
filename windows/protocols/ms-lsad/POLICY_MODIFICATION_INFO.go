package mslsad

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// POLICY_MODIFICATION_INFO is an obsolete policy information class ([MS-LSAD] 2.2.4.10).
type POLICY_MODIFICATION_INFO struct {
	ModifiedId           msdtyp.LARGE_INTEGER
	DatabaseCreationTime msdtyp.LARGE_INTEGER
}
