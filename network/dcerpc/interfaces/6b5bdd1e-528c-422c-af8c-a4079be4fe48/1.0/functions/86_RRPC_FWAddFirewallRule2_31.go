package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWAddFirewallRule2_31Request carries the [in] parameters of RRPC_FWAddFirewallRule2_31.
type rRPC_FWAddFirewallRule2_31Request struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PRule        msfasp.FW_RULE2_31
}

func (*rRPC_FWAddFirewallRule2_31Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWAddFirewallRule2_31
}

// rRPC_FWAddFirewallRule2_31Response carries the [out] parameters and return value of RRPC_FWAddFirewallRule2_31.
type rRPC_FWAddFirewallRule2_31Response struct {
	PStatus msfasp.FW_RULE_STATUS
	Status  ndr.DWORD `ndr:"retval"`
}

// RRPC_FWAddFirewallRule2_31 calls RRPC_FWAddFirewallRule2_31 (opnum 86) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWAddFirewallRule2_31(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pRule msfasp.FW_RULE2_31) (PStatus msfasp.FW_RULE_STATUS, err error) {
	req := &rRPC_FWAddFirewallRule2_31Request{
		HPolicyStore: hPolicyStore,
		PRule:        pRule,
	}
	var resp rRPC_FWAddFirewallRule2_31Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWAddFirewallRule2_31: %w", err)
		return
	}
	PStatus = resp.PStatus
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWAddFirewallRule2_31 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
