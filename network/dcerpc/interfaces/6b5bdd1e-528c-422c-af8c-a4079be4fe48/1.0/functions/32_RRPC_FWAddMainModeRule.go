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

// rRPC_FWAddMainModeRuleRequest carries the [in] parameters of RRPC_FWAddMainModeRule.
type rRPC_FWAddMainModeRuleRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PMMRule      msfasp.FW_MM_RULE
}

func (*rRPC_FWAddMainModeRuleRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWAddMainModeRule }

// rRPC_FWAddMainModeRuleResponse carries the [out] parameters and return value of RRPC_FWAddMainModeRule.
type rRPC_FWAddMainModeRuleResponse struct {
	PStatus msfasp.FW_RULE_STATUS
	Status  ndr.DWORD `ndr:"retval"`
}

// RRPC_FWAddMainModeRule calls RRPC_FWAddMainModeRule (opnum 32) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWAddMainModeRule(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pMMRule msfasp.FW_MM_RULE) (PStatus msfasp.FW_RULE_STATUS, err error) {
	req := &rRPC_FWAddMainModeRuleRequest{
		HPolicyStore: hPolicyStore,
		PMMRule:      pMMRule,
	}
	var resp rRPC_FWAddMainModeRuleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWAddMainModeRule: %w", err)
		return
	}
	PStatus = resp.PStatus
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWAddMainModeRule failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
