package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWQueryFirewallRules2_27Request carries the [in] parameters of RRPC_FWQueryFirewallRules2_27.
type rRPC_FWQueryFirewallRules2_27Request struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PQuery       msfasp.FW_QUERY
	WFlags       uint16
}

func (*rRPC_FWQueryFirewallRules2_27Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWQueryFirewallRules2_27
}

// rRPC_FWQueryFirewallRules2_27Response carries the [out] parameters and return value of RRPC_FWQueryFirewallRules2_27.
type rRPC_FWQueryFirewallRules2_27Response struct {
	PdwNumRules ndr.DWORD
	PpRules     *msfasp.FW_RULE2_27 `ndr:"unique"`
	Status      ndr.DWORD           `ndr:"retval"`
}

// RRPC_FWQueryFirewallRules2_27 calls RRPC_FWQueryFirewallRules2_27 (opnum 85) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWQueryFirewallRules2_27(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pQuery msfasp.FW_QUERY, wFlags uint16) (PdwNumRules ndr.DWORD, PpRules *msfasp.FW_RULE2_27, err error) {
	req := &rRPC_FWQueryFirewallRules2_27Request{
		HPolicyStore: hPolicyStore,
		PQuery:       pQuery,
		WFlags:       wFlags,
	}
	var resp rRPC_FWQueryFirewallRules2_27Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWQueryFirewallRules2_27: %w", err)
		return
	}
	PdwNumRules = resp.PdwNumRules
	PpRules = resp.PpRules
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWQueryFirewallRules2_27 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
