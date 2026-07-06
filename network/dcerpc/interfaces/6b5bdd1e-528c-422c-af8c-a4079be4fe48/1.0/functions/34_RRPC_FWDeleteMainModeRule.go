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

// rRPC_FWDeleteMainModeRuleRequest carries the [in] parameters of RRPC_FWDeleteMainModeRule.
type rRPC_FWDeleteMainModeRuleRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PRuleId      ndr.WSTR
}

func (*rRPC_FWDeleteMainModeRuleRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWDeleteMainModeRule
}

// rRPC_FWDeleteMainModeRuleResponse carries the [out] parameters and return value of RRPC_FWDeleteMainModeRule.
type rRPC_FWDeleteMainModeRuleResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWDeleteMainModeRule calls RRPC_FWDeleteMainModeRule (opnum 34) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWDeleteMainModeRule(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pRuleId ndr.WSTR) (err error) {
	req := &rRPC_FWDeleteMainModeRuleRequest{
		HPolicyStore: hPolicyStore,
		PRuleId:      pRuleId,
	}
	var resp rRPC_FWDeleteMainModeRuleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWDeleteMainModeRule: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWDeleteMainModeRule failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
