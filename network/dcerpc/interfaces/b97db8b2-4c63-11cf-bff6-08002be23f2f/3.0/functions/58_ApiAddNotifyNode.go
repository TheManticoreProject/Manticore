package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiAddNotifyNodeRequest carries the [in] parameters of ApiAddNotifyNode.
type apiAddNotifyNodeRequest struct {
	HNotify     mscmrp.HNOTIFY_RPC
	HNode       mscmrp.HNODE_RPC
	DwFilter    ndr.DWORD
	DwNotifyKey ndr.DWORD
}

func (*apiAddNotifyNodeRequest) Opnum() uint16 { return clusapi.OpnumApiAddNotifyNode }

// apiAddNotifyNodeResponse carries the [out] parameters and return value of ApiAddNotifyNode.
type apiAddNotifyNodeResponse struct {
	DwStateSequence ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiAddNotifyNode calls ApiAddNotifyNode (opnum 58) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddNotifyNode(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hNode mscmrp.HNODE_RPC, dwFilter ndr.DWORD, dwNotifyKey ndr.DWORD) (DwStateSequence ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiAddNotifyNodeRequest{
		HNotify:     hNotify,
		HNode:       hNode,
		DwFilter:    dwFilter,
		DwNotifyKey: dwNotifyKey,
	}
	var resp apiAddNotifyNodeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddNotifyNode: %w", err)
		return
	}
	DwStateSequence = resp.DwStateSequence
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddNotifyNode failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
