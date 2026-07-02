package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWEnumPhase2SAsRequest carries the [in] parameters of RRPC_FWEnumPhase2SAs.
type rRPC_FWEnumPhase2SAsRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PEndpoints   *msfasp.FW_ENDPOINTS `ndr:"unique"`
}

func (*rRPC_FWEnumPhase2SAsRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWEnumPhase2SAs }

// rRPC_FWEnumPhase2SAsResponse carries the [out] parameters and return value of RRPC_FWEnumPhase2SAs.
type rRPC_FWEnumPhase2SAsResponse struct {
	PdwNumSAs ndr.DWORD
	PpSAs     []*msfasp.FW_PHASE2_SA_DETAILS `ndr:"elem=unique,ref,conformant"`
	Status    ndr.DWORD                      `ndr:"retval"`
}

// RRPC_FWEnumPhase2SAs calls RRPC_FWEnumPhase2SAs (opnum 28) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumPhase2SAs(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pEndpoints *msfasp.FW_ENDPOINTS) (PdwNumSAs ndr.DWORD, PpSAs []*msfasp.FW_PHASE2_SA_DETAILS, err error) {
	req := &rRPC_FWEnumPhase2SAsRequest{
		HPolicyStore: hPolicyStore,
		PEndpoints:   pEndpoints,
	}
	var resp rRPC_FWEnumPhase2SAsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumPhase2SAs: %w", err)
		return
	}
	PdwNumSAs = resp.PdwNumSAs
	PpSAs = resp.PpSAs
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumPhase2SAs failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
