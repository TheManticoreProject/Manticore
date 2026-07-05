package msscmr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SC_RPC_NOTIFY_PARAMS_LIST ([MS-SCMR] 2.2.23). NotifyParamsArray is declared in the IDL as
// a conformant array ("[size_is(cElements)] SC_RPC_NOTIFY_PARAMS NotifyParamsArray[*]"), i.e.
// embedded inline (its maximum_count is hoisted ahead of the struct) — not a [unique] pointer.
type SC_RPC_NOTIFY_PARAMS_LIST struct {
	CElements         ndr.DWORD
	NotifyParamsArray []SC_RPC_NOTIFY_PARAMS `ndr:"conformant,size_is=CElements"`
}
