package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCreateGroupExRequest carries the [in] parameters of ApiCreateGroupEx.
type apiCreateGroupExRequest struct {
	LpszGroupName ndr.WSTR
	PGroupInfo    *mscmrp.CLUSTER_CREATE_GROUP_INFO_RPC `ndr:"unique"`
}

func (*apiCreateGroupExRequest) Opnum() uint16 { return clusapi.OpnumApiCreateGroupEx }

// apiCreateGroupExResponse carries the [out] parameters and return value of ApiCreateGroupEx.
type apiCreateGroupExResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HGROUP_RPC `ndr:"retval"`
}

// ApiCreateGroupEx calls ApiCreateGroupEx (opnum 129) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateGroupEx(rpc ndr.Invoker, lpszGroupName ndr.WSTR, pGroupInfo *mscmrp.CLUSTER_CREATE_GROUP_INFO_RPC) (Handle mscmrp.HGROUP_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateGroupExRequest{
		LpszGroupName: lpszGroupName,
		PGroupInfo:    pGroupInfo,
	}
	var resp apiCreateGroupExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateGroupEx: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateGroupEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
