package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWEnumFirewallRules2_20Request carries the [in] parameters of RRPC_FWEnumFirewallRules2_20.
type rRPC_FWEnumFirewallRules2_20Request struct {
	HPolicyStore       msfasp.FW_POLICY_STORE_HANDLE
	DwFilteredByStatus ndr.DWORD
	DwProfileFilter    ndr.DWORD
	WFlags             uint16
}

func (*rRPC_FWEnumFirewallRules2_20Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWEnumFirewallRules2_20
}

// rRPC_FWEnumFirewallRules2_20Response carries the [out] parameters and return value of RRPC_FWEnumFirewallRules2_20.
type rRPC_FWEnumFirewallRules2_20Response struct {
	PdwNumRules ndr.DWORD
	PpRules     *msfasp.FW_RULE2_20 `ndr:"unique"`
	Status      ndr.DWORD           `ndr:"retval"`
}

// RRPC_FWEnumFirewallRules2_20 calls RRPC_FWEnumFirewallRules2_20 (opnum 68) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumFirewallRules2_20(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, dwFilteredByStatus ndr.DWORD, dwProfileFilter ndr.DWORD, wFlags uint16) (PdwNumRules ndr.DWORD, PpRules *msfasp.FW_RULE2_20, err error) {
	req := &rRPC_FWEnumFirewallRules2_20Request{
		HPolicyStore:       hPolicyStore,
		DwFilteredByStatus: dwFilteredByStatus,
		DwProfileFilter:    dwProfileFilter,
		WFlags:             wFlags,
	}
	var resp rRPC_FWEnumFirewallRules2_20Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumFirewallRules2_20: %w", err)
		return
	}
	PdwNumRules = resp.PdwNumRules
	PpRules = resp.PpRules
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumFirewallRules2_20 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
