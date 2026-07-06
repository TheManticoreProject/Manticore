package functions

// IDL source: [MS-FASP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fasp/1503b9d7-7fec-4793-9972-6ad58720c9db
// A fetched copy is kept at ms-fasp.idl in the interface directory.

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWDeleteFirewallRuleRequest carries the [in] parameters of RRPC_FWDeleteFirewallRule.
type rRPC_FWDeleteFirewallRuleRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	WszRuleID    ndr.WSTR
}

func (*rRPC_FWDeleteFirewallRuleRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWDeleteFirewallRule
}

// rRPC_FWDeleteFirewallRuleResponse carries the [out] parameters and return value of RRPC_FWDeleteFirewallRule.
type rRPC_FWDeleteFirewallRuleResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWDeleteFirewallRule calls RRPC_FWDeleteFirewallRule (opnum 7) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWDeleteFirewallRule(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, wszRuleID ndr.WSTR) (err error) {
	req := &rRPC_FWDeleteFirewallRuleRequest{
		HPolicyStore: hPolicyStore,
		WszRuleID:    wszRuleID,
	}
	var resp rRPC_FWDeleteFirewallRuleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWDeleteFirewallRule: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWDeleteFirewallRule failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
