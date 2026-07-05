package msscmr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SC_RPC_SERVICE_CONTROL_OUT_PARAMSW ([MS-SCMR] 2.2.26) is a [switch_type(DWORD)] union; its
// single IDL arm is a [unique] pointer (P*), so it emits a referent id with its body deferred
// ([C706] 14.3.8). Tag is the inline 4-byte discriminant.
type SC_RPC_SERVICE_CONTROL_OUT_PARAMSW struct {
	Tag          ndr.DWORD                                 `ndr:"switch"`
	PsrOutParams *SERVICE_CONTROL_STATUS_REASON_OUT_PARAMS `ndr:"case=1,unique"`
}
