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

// rRPC_FWAddFirewallRule2_27Request carries the [in] parameters of RRPC_FWAddFirewallRule2_27.
type rRPC_FWAddFirewallRule2_27Request struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PRule        msfasp.FW_RULE2_27
}

func (*rRPC_FWAddFirewallRule2_27Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWAddFirewallRule2_27
}

// rRPC_FWAddFirewallRule2_27Response carries the [out] parameters and return value of RRPC_FWAddFirewallRule2_27.
type rRPC_FWAddFirewallRule2_27Response struct {
	PStatus msfasp.FW_RULE_STATUS
	Status  ndr.DWORD `ndr:"retval"`
}

// RRPC_FWAddFirewallRule2_27 calls RRPC_FWAddFirewallRule2_27 (opnum 82) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWAddFirewallRule2_27(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pRule msfasp.FW_RULE2_27) (PStatus msfasp.FW_RULE_STATUS, err error) {
	req := &rRPC_FWAddFirewallRule2_27Request{
		HPolicyStore: hPolicyStore,
		PRule:        pRule,
	}
	var resp rRPC_FWAddFirewallRule2_27Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWAddFirewallRule2_27: %w", err)
		return
	}
	PStatus = resp.PStatus
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWAddFirewallRule2_27 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
