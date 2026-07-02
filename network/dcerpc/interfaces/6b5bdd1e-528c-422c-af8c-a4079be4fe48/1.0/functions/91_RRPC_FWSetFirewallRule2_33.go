package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWSetFirewallRule2_33Request carries the [in] parameters of RRPC_FWSetFirewallRule2_33.
type rRPC_FWSetFirewallRule2_33Request struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PRule        msfasp.FW_RULE
}

func (*rRPC_FWSetFirewallRule2_33Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWSetFirewallRule2_33
}

// rRPC_FWSetFirewallRule2_33Response carries the [out] parameters and return value of RRPC_FWSetFirewallRule2_33.
type rRPC_FWSetFirewallRule2_33Response struct {
	PStatus msfasp.FW_RULE_STATUS
	Status  ndr.DWORD `ndr:"retval"`
}

// RRPC_FWSetFirewallRule2_33 calls RRPC_FWSetFirewallRule2_33 (opnum 91) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWSetFirewallRule2_33(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pRule msfasp.FW_RULE) (PStatus msfasp.FW_RULE_STATUS, err error) {
	req := &rRPC_FWSetFirewallRule2_33Request{
		HPolicyStore: hPolicyStore,
		PRule:        pRule,
	}
	var resp rRPC_FWSetFirewallRule2_33Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWSetFirewallRule2_33: %w", err)
		return
	}
	PStatus = resp.PStatus
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWSetFirewallRule2_33 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
