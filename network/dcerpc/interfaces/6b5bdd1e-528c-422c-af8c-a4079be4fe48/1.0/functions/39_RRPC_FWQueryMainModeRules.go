package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWQueryMainModeRulesRequest carries the [in] parameters of RRPC_FWQueryMainModeRules.
type rRPC_FWQueryMainModeRulesRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PQuery       msfasp.FW_QUERY
	WFlags       uint16
}

func (*rRPC_FWQueryMainModeRulesRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWQueryMainModeRules
}

// rRPC_FWQueryMainModeRulesResponse carries the [out] parameters and return value of RRPC_FWQueryMainModeRules.
type rRPC_FWQueryMainModeRulesResponse struct {
	PdwNumRules ndr.DWORD
	PpMMRules   *msfasp.FW_MM_RULE `ndr:"unique"`
	Status      ndr.DWORD          `ndr:"retval"`
}

// RRPC_FWQueryMainModeRules calls RRPC_FWQueryMainModeRules (opnum 39) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWQueryMainModeRules(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pQuery msfasp.FW_QUERY, wFlags uint16) (PdwNumRules ndr.DWORD, PpMMRules *msfasp.FW_MM_RULE, err error) {
	req := &rRPC_FWQueryMainModeRulesRequest{
		HPolicyStore: hPolicyStore,
		PQuery:       pQuery,
		WFlags:       wFlags,
	}
	var resp rRPC_FWQueryMainModeRulesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWQueryMainModeRules: %w", err)
		return
	}
	PdwNumRules = resp.PdwNumRules
	PpMMRules = resp.PpMMRules
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWQueryMainModeRules failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
