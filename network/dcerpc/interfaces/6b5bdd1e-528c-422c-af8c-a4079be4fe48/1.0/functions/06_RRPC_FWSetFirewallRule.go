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

// rRPC_FWSetFirewallRuleRequest carries the [in] parameters of RRPC_FWSetFirewallRule.
type rRPC_FWSetFirewallRuleRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PRule        msfasp.FW_RULE2_0
}

func (*rRPC_FWSetFirewallRuleRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWSetFirewallRule }

// rRPC_FWSetFirewallRuleResponse carries the [out] parameters and return value of RRPC_FWSetFirewallRule.
type rRPC_FWSetFirewallRuleResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWSetFirewallRule calls RRPC_FWSetFirewallRule (opnum 6) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWSetFirewallRule(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pRule msfasp.FW_RULE2_0) (err error) {
	req := &rRPC_FWSetFirewallRuleRequest{
		HPolicyStore: hPolicyStore,
		PRule:        pRule,
	}
	var resp rRPC_FWSetFirewallRuleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWSetFirewallRule: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWSetFirewallRule failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
