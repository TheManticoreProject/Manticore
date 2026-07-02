package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWEnumPhase1SAsRequest carries the [in] parameters of RRPC_FWEnumPhase1SAs.
type rRPC_FWEnumPhase1SAsRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PEndpoints   *msfasp.FW_ENDPOINTS `ndr:"unique"`
}

func (*rRPC_FWEnumPhase1SAsRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWEnumPhase1SAs }

// rRPC_FWEnumPhase1SAsResponse carries the [out] parameters and return value of RRPC_FWEnumPhase1SAs.
type rRPC_FWEnumPhase1SAsResponse struct {
	PdwNumSAs ndr.DWORD
	PpSAs     []*msfasp.FW_PHASE1_SA_DETAILS `ndr:"elem=unique,ref,conformant"`
	Status    ndr.DWORD                      `ndr:"retval"`
}

// RRPC_FWEnumPhase1SAs calls RRPC_FWEnumPhase1SAs (opnum 27) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumPhase1SAs(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pEndpoints *msfasp.FW_ENDPOINTS) (PdwNumSAs ndr.DWORD, PpSAs []*msfasp.FW_PHASE1_SA_DETAILS, err error) {
	req := &rRPC_FWEnumPhase1SAsRequest{
		HPolicyStore: hPolicyStore,
		PEndpoints:   pEndpoints,
	}
	var resp rRPC_FWEnumPhase1SAsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumPhase1SAs: %w", err)
		return
	}
	PdwNumSAs = resp.PdwNumSAs
	PpSAs = resp.PpSAs
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumPhase1SAs failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
