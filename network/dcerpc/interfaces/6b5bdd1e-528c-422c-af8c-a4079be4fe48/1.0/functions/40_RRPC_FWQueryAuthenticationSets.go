package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWQueryAuthenticationSetsRequest carries the [in] parameters of RRPC_FWQueryAuthenticationSets.
type rRPC_FWQueryAuthenticationSetsRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	IPsecPhase   msfasp.FW_IPSEC_PHASE
	PQuery       msfasp.FW_QUERY
	WFlags       uint16
}

func (*rRPC_FWQueryAuthenticationSetsRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWQueryAuthenticationSets
}

// rRPC_FWQueryAuthenticationSetsResponse carries the [out] parameters and return value of RRPC_FWQueryAuthenticationSets.
type rRPC_FWQueryAuthenticationSetsResponse struct {
	PdwNumSets ndr.DWORD
	PpAuthSets *msfasp.FW_AUTH_SET2_10 `ndr:"unique"`
	Status     ndr.DWORD               `ndr:"retval"`
}

// RRPC_FWQueryAuthenticationSets calls RRPC_FWQueryAuthenticationSets (opnum 40) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWQueryAuthenticationSets(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, iPsecPhase msfasp.FW_IPSEC_PHASE, pQuery msfasp.FW_QUERY, wFlags uint16) (PdwNumSets ndr.DWORD, PpAuthSets *msfasp.FW_AUTH_SET2_10, err error) {
	req := &rRPC_FWQueryAuthenticationSetsRequest{
		HPolicyStore: hPolicyStore,
		IPsecPhase:   iPsecPhase,
		PQuery:       pQuery,
		WFlags:       wFlags,
	}
	var resp rRPC_FWQueryAuthenticationSetsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWQueryAuthenticationSets: %w", err)
		return
	}
	PdwNumSets = resp.PdwNumSets
	PpAuthSets = resp.PpAuthSets
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWQueryAuthenticationSets failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
