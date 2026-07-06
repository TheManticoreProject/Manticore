package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiPauseNodeExRequest carries the [in] parameters of ApiPauseNodeEx.
type apiPauseNodeExRequest struct {
	HNode        mscmrp.HNODE_RPC
	BDrainNode   ndr.BOOL
	DwPauseFlags ndr.DWORD
}

func (*apiPauseNodeExRequest) Opnum() uint16 { return clusapi.OpnumApiPauseNodeEx }

// apiPauseNodeExResponse carries the [out] parameters and return value of ApiPauseNodeEx.
type apiPauseNodeExResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiPauseNodeEx calls ApiPauseNodeEx (opnum 126) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiPauseNodeEx(rpc ndr.Invoker, hNode mscmrp.HNODE_RPC, bDrainNode ndr.BOOL, dwPauseFlags ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiPauseNodeExRequest{
		HNode:        hNode,
		BDrainNode:   bDrainNode,
		DwPauseFlags: dwPauseFlags,
	}
	var resp apiPauseNodeExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiPauseNodeEx: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiPauseNodeEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
