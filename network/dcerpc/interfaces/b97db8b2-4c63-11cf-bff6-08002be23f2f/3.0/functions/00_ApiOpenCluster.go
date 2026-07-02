package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOpenClusterRequest carries the [in] parameters of ApiOpenCluster.
type apiOpenClusterRequest struct {
}

func (*apiOpenClusterRequest) Opnum() uint16 { return clusapi.OpnumApiOpenCluster }

// apiOpenClusterResponse carries the [out] parameters and return value of ApiOpenCluster.
type apiOpenClusterResponse struct {
	Status ndr.DWORD
	Handle mscmrp.HCLUSTER_RPC `ndr:"retval"`
}

// ApiOpenCluster calls ApiOpenCluster (opnum 0) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenCluster(rpc ndr.Invoker) (Handle mscmrp.HCLUSTER_RPC, Status ndr.DWORD, err error) {
	req := &apiOpenClusterRequest{}
	var resp apiOpenClusterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenCluster: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenCluster failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
