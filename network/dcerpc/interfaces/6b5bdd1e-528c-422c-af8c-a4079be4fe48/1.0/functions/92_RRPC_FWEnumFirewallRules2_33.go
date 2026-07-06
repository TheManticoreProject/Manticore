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

// rRPC_FWEnumFirewallRules2_33Request carries the [in] parameters of RRPC_FWEnumFirewallRules2_33.
type rRPC_FWEnumFirewallRules2_33Request struct {
	HPolicyStore       msfasp.FW_POLICY_STORE_HANDLE
	DwFilteredByStatus ndr.DWORD
	DwProfileFilter    ndr.DWORD
	WFlags             uint16
}

func (*rRPC_FWEnumFirewallRules2_33Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWEnumFirewallRules2_33
}

// rRPC_FWEnumFirewallRules2_33Response carries the [out] parameters and return value of RRPC_FWEnumFirewallRules2_33.
type rRPC_FWEnumFirewallRules2_33Response struct {
	PdwNumRules ndr.DWORD
	PpRules     *msfasp.FW_RULE `ndr:"unique"`
	Status      ndr.DWORD       `ndr:"retval"`
}

// RRPC_FWEnumFirewallRules2_33 calls RRPC_FWEnumFirewallRules2_33 (opnum 92) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumFirewallRules2_33(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, dwFilteredByStatus ndr.DWORD, dwProfileFilter ndr.DWORD, wFlags uint16) (PdwNumRules ndr.DWORD, PpRules *msfasp.FW_RULE, err error) {
	req := &rRPC_FWEnumFirewallRules2_33Request{
		HPolicyStore:       hPolicyStore,
		DwFilteredByStatus: dwFilteredByStatus,
		DwProfileFilter:    dwProfileFilter,
		WFlags:             wFlags,
	}
	var resp rRPC_FWEnumFirewallRules2_33Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumFirewallRules2_33: %w", err)
		return
	}
	PdwNumRules = resp.PdwNumRules
	PpRules = resp.PpRules
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumFirewallRules2_33 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
