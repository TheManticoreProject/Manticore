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

// apiMoveGroupToNodeExRequest carries the [in] parameters of ApiMoveGroupToNodeEx.
type apiMoveGroupToNodeExRequest struct {
	HGroup         mscmrp.HGROUP_RPC
	HNode          mscmrp.HNODE_RPC
	DwMoveFlags    ndr.DWORD
	LpInBuffer     []uint8 `ndr:"ref,size_is=CbInBufferSize"`
	CbInBufferSize ndr.DWORD
}

func (*apiMoveGroupToNodeExRequest) Opnum() uint16 { return clusapi.OpnumApiMoveGroupToNodeEx }

// apiMoveGroupToNodeExResponse carries the [out] parameters and return value of ApiMoveGroupToNodeEx.
type apiMoveGroupToNodeExResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiMoveGroupToNodeEx calls ApiMoveGroupToNodeEx (opnum 133) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiMoveGroupToNodeEx(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, hNode mscmrp.HNODE_RPC, dwMoveFlags ndr.DWORD, lpInBuffer []uint8, cbInBufferSize ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiMoveGroupToNodeExRequest{
		HGroup:         hGroup,
		HNode:          hNode,
		DwMoveFlags:    dwMoveFlags,
		LpInBuffer:     lpInBuffer,
		CbInBufferSize: cbInBufferSize,
	}
	var resp apiMoveGroupToNodeExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiMoveGroupToNodeEx: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiMoveGroupToNodeEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
