package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWDeleteAllMainModeRulesRequest carries the [in] parameters of RRPC_FWDeleteAllMainModeRules.
type rRPC_FWDeleteAllMainModeRulesRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
}

func (*rRPC_FWDeleteAllMainModeRulesRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWDeleteAllMainModeRules
}

// rRPC_FWDeleteAllMainModeRulesResponse carries the [out] parameters and return value of RRPC_FWDeleteAllMainModeRules.
type rRPC_FWDeleteAllMainModeRulesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWDeleteAllMainModeRules calls RRPC_FWDeleteAllMainModeRules (opnum 35) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWDeleteAllMainModeRules(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE) (err error) {
	req := &rRPC_FWDeleteAllMainModeRulesRequest{
		HPolicyStore: hPolicyStore,
	}
	var resp rRPC_FWDeleteAllMainModeRulesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWDeleteAllMainModeRules: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWDeleteAllMainModeRules failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
