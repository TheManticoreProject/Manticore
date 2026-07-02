package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiAddNotifyClusterRequest carries the [in] parameters of ApiAddNotifyCluster.
type apiAddNotifyClusterRequest struct {
	HNotify     mscmrp.HNOTIFY_RPC
	HCluster    mscmrp.HCLUSTER_RPC
	DwFilter    ndr.DWORD
	DwNotifyKey ndr.DWORD
}

func (*apiAddNotifyClusterRequest) Opnum() uint16 { return clusapi.OpnumApiAddNotifyCluster }

// apiAddNotifyClusterResponse carries the [out] parameters and return value of ApiAddNotifyCluster.
type apiAddNotifyClusterResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiAddNotifyCluster calls ApiAddNotifyCluster (opnum 57) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddNotifyCluster(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hCluster mscmrp.HCLUSTER_RPC, dwFilter ndr.DWORD, dwNotifyKey ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiAddNotifyClusterRequest{
		HNotify:     hNotify,
		HCluster:    hCluster,
		DwFilter:    dwFilter,
		DwNotifyKey: dwNotifyKey,
	}
	var resp apiAddNotifyClusterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddNotifyCluster: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddNotifyCluster failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
