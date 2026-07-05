package msscmr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SC_RPC_SERVICE_CONTROL_IN_PARAMSA ([MS-SCMR] 2.2.24) is a [switch_type(DWORD)] union; its
// single IDL arm is a [unique] pointer (P*), so it emits a referent id with its body deferred
// ([C706] 14.3.8). Tag is the inline 4-byte discriminant.
type SC_RPC_SERVICE_CONTROL_IN_PARAMSA struct {
	Tag         ndr.DWORD                                 `ndr:"switch"`
	PsrInParams *SERVICE_CONTROL_STATUS_REASON_IN_PARAMSA `ndr:"case=1,unique"`
}
