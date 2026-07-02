package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWDeleteAllFirewallRulesRequest carries the [in] parameters of RRPC_FWDeleteAllFirewallRules.
type rRPC_FWDeleteAllFirewallRulesRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
}

func (*rRPC_FWDeleteAllFirewallRulesRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWDeleteAllFirewallRules
}

// rRPC_FWDeleteAllFirewallRulesResponse carries the [out] parameters and return value of RRPC_FWDeleteAllFirewallRules.
type rRPC_FWDeleteAllFirewallRulesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWDeleteAllFirewallRules calls RRPC_FWDeleteAllFirewallRules (opnum 8) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWDeleteAllFirewallRules(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE) (err error) {
	req := &rRPC_FWDeleteAllFirewallRulesRequest{
		HPolicyStore: hPolicyStore,
	}
	var resp rRPC_FWDeleteAllFirewallRulesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWDeleteAllFirewallRules: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWDeleteAllFirewallRules failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
