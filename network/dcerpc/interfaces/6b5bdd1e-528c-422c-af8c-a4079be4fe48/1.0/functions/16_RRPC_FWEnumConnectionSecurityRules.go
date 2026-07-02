package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWEnumConnectionSecurityRulesRequest carries the [in] parameters of RRPC_FWEnumConnectionSecurityRules.
type rRPC_FWEnumConnectionSecurityRulesRequest struct {
	HPolicyStore       msfasp.FW_POLICY_STORE_HANDLE
	DwFilteredByStatus ndr.DWORD
	DwProfileFilter    ndr.DWORD
	WFlags             uint16
}

func (*rRPC_FWEnumConnectionSecurityRulesRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWEnumConnectionSecurityRules
}

// rRPC_FWEnumConnectionSecurityRulesResponse carries the [out] parameters and return value of RRPC_FWEnumConnectionSecurityRules.
type rRPC_FWEnumConnectionSecurityRulesResponse struct {
	PdwNumRules ndr.DWORD
	PpRules     *msfasp.FW_CS_RULE2_0 `ndr:"unique"`
	Status      ndr.DWORD             `ndr:"retval"`
}

// RRPC_FWEnumConnectionSecurityRules calls RRPC_FWEnumConnectionSecurityRules (opnum 16) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumConnectionSecurityRules(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, dwFilteredByStatus ndr.DWORD, dwProfileFilter ndr.DWORD, wFlags uint16) (PdwNumRules ndr.DWORD, PpRules *msfasp.FW_CS_RULE2_0, err error) {
	req := &rRPC_FWEnumConnectionSecurityRulesRequest{
		HPolicyStore:       hPolicyStore,
		DwFilteredByStatus: dwFilteredByStatus,
		DwProfileFilter:    dwProfileFilter,
		WFlags:             wFlags,
	}
	var resp rRPC_FWEnumConnectionSecurityRulesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumConnectionSecurityRules: %w", err)
		return
	}
	PdwNumRules = resp.PdwNumRules
	PpRules = resp.PpRules
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumConnectionSecurityRules failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
