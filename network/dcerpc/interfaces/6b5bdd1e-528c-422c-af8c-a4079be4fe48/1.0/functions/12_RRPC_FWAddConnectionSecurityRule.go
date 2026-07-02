package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWAddConnectionSecurityRuleRequest carries the [in] parameters of RRPC_FWAddConnectionSecurityRule.
type rRPC_FWAddConnectionSecurityRuleRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PRule        msfasp.FW_CS_RULE2_0
}

func (*rRPC_FWAddConnectionSecurityRuleRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWAddConnectionSecurityRule
}

// rRPC_FWAddConnectionSecurityRuleResponse carries the [out] parameters and return value of RRPC_FWAddConnectionSecurityRule.
type rRPC_FWAddConnectionSecurityRuleResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWAddConnectionSecurityRule calls RRPC_FWAddConnectionSecurityRule (opnum 12) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWAddConnectionSecurityRule(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pRule msfasp.FW_CS_RULE2_0) (err error) {
	req := &rRPC_FWAddConnectionSecurityRuleRequest{
		HPolicyStore: hPolicyStore,
		PRule:        pRule,
	}
	var resp rRPC_FWAddConnectionSecurityRuleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWAddConnectionSecurityRule: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWAddConnectionSecurityRule failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
