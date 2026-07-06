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

// rRPC_FWSetMainModeRuleRequest carries the [in] parameters of RRPC_FWSetMainModeRule.
type rRPC_FWSetMainModeRuleRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PMMRule      msfasp.FW_MM_RULE
}

func (*rRPC_FWSetMainModeRuleRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWSetMainModeRule }

// rRPC_FWSetMainModeRuleResponse carries the [out] parameters and return value of RRPC_FWSetMainModeRule.
type rRPC_FWSetMainModeRuleResponse struct {
	PStatus msfasp.FW_RULE_STATUS
	Status  ndr.DWORD `ndr:"retval"`
}

// RRPC_FWSetMainModeRule calls RRPC_FWSetMainModeRule (opnum 33) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWSetMainModeRule(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pMMRule msfasp.FW_MM_RULE) (PStatus msfasp.FW_RULE_STATUS, err error) {
	req := &rRPC_FWSetMainModeRuleRequest{
		HPolicyStore: hPolicyStore,
		PMMRule:      pMMRule,
	}
	var resp rRPC_FWSetMainModeRuleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWSetMainModeRule: %w", err)
		return
	}
	PStatus = resp.PStatus
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWSetMainModeRule failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
