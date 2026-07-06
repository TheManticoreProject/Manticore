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

// apiCloseClusterRequest carries the [in] parameters of ApiCloseCluster.
type apiCloseClusterRequest struct {
	Cluster mscmrp.HCLUSTER_RPC
}

func (*apiCloseClusterRequest) Opnum() uint16 { return clusapi.OpnumApiCloseCluster }

// apiCloseClusterResponse carries the [out] parameters and return value of ApiCloseCluster.
type apiCloseClusterResponse struct {
	Cluster mscmrp.HCLUSTER_RPC
	Status  ndr.DWORD `ndr:"retval"`
}

// ApiCloseCluster calls ApiCloseCluster (opnum 1) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCloseCluster(rpc ndr.Invoker, cluster mscmrp.HCLUSTER_RPC) (Cluster mscmrp.HCLUSTER_RPC, err error) {
	req := &apiCloseClusterRequest{
		Cluster: cluster,
	}
	var resp apiCloseClusterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCloseCluster: %w", err)
		return
	}
	Cluster = resp.Cluster
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCloseCluster failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
