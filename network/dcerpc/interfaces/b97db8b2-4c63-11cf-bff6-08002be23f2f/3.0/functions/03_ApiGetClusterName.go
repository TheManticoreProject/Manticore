package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

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
