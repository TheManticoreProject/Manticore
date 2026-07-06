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

// rRPC_FWQueryFirewallRules2_31Request carries the [in] parameters of RRPC_FWQueryFirewallRules2_31.
type rRPC_FWQueryFirewallRules2_31Request struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PQuery       msfasp.FW_QUERY
	WFlags       uint16
}

func (*rRPC_FWQueryFirewallRules2_31Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWQueryFirewallRules2_31
}

// rRPC_FWQueryFirewallRules2_31Response carries the [out] parameters and return value of RRPC_FWQueryFirewallRules2_31.
type rRPC_FWQueryFirewallRules2_31Response struct {
	PdwNumRules ndr.DWORD
	PpRules     *msfasp.FW_RULE2_31 `ndr:"unique"`
	Status      ndr.DWORD           `ndr:"retval"`
}

// RRPC_FWQueryFirewallRules2_31 calls RRPC_FWQueryFirewallRules2_31 (opnum 89) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWQueryFirewallRules2_31(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pQuery msfasp.FW_QUERY, wFlags uint16) (PdwNumRules ndr.DWORD, PpRules *msfasp.FW_RULE2_31, err error) {
	req := &rRPC_FWQueryFirewallRules2_31Request{
		HPolicyStore: hPolicyStore,
		PQuery:       pQuery,
		WFlags:       wFlags,
	}
	var resp rRPC_FWQueryFirewallRules2_31Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWQueryFirewallRules2_31: %w", err)
		return
	}
	PdwNumRules = resp.PdwNumRules
	PpRules = resp.PpRules
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWQueryFirewallRules2_31 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
