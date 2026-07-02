package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWEnumMainModeRulesRequest carries the [in] parameters of RRPC_FWEnumMainModeRules.
type rRPC_FWEnumMainModeRulesRequest struct {
	HPolicyStore       msfasp.FW_POLICY_STORE_HANDLE
	DwFilteredByStatus ndr.DWORD
	DwProfileFilter    ndr.DWORD
	WFlags             uint16
}

func (*rRPC_FWEnumMainModeRulesRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWEnumMainModeRules }

// rRPC_FWEnumMainModeRulesResponse carries the [out] parameters and return value of RRPC_FWEnumMainModeRules.
type rRPC_FWEnumMainModeRulesResponse struct {
	PdwNumRules ndr.DWORD
	PpMMRules   *msfasp.FW_MM_RULE `ndr:"unique"`
	Status      ndr.DWORD          `ndr:"retval"`
}

// RRPC_FWEnumMainModeRules calls RRPC_FWEnumMainModeRules (opnum 36) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumMainModeRules(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, dwFilteredByStatus ndr.DWORD, dwProfileFilter ndr.DWORD, wFlags uint16) (PdwNumRules ndr.DWORD, PpMMRules *msfasp.FW_MM_RULE, err error) {
	req := &rRPC_FWEnumMainModeRulesRequest{
		HPolicyStore:       hPolicyStore,
		DwFilteredByStatus: dwFilteredByStatus,
		DwProfileFilter:    dwProfileFilter,
		WFlags:             wFlags,
	}
	var resp rRPC_FWEnumMainModeRulesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumMainModeRules: %w", err)
		return
	}
	PdwNumRules = resp.PdwNumRules
	PpMMRules = resp.PpMMRules
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumMainModeRules failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
