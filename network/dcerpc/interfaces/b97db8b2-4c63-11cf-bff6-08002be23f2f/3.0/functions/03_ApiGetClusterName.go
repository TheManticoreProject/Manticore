package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// apiGetClusterNameRequest carries the [in] parameters of ApiGetClusterName.
type apiGetClusterNameRequest struct {
}

func (*apiGetClusterNameRequest) Opnum() uint16 { return clusapi.OpnumApiGetClusterName }

// apiGetClusterNameResponse carries the [out] parameters and return value of ApiGetClusterName.
type apiGetClusterNameResponse struct {
	ClusterName ndr.WSTR
	NodeName    ndr.WSTR
	Status      ndr.DWORD `ndr:"retval"`
}

// ApiGetClusterName calls ApiGetClusterName (opnum 3) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetClusterName(rpc ndr.Invoker) (ClusterName ndr.WSTR, NodeName ndr.WSTR, err error) {
	req := &apiGetClusterNameRequest{}
	var resp apiGetClusterNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetClusterName: %w", err)
		return
	}
	ClusterName = resp.ClusterName
	NodeName = resp.NodeName
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetClusterName failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
