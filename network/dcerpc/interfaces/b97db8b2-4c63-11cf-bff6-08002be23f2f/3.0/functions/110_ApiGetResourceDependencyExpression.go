package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetResourceDependencyExpressionRequest carries the [in] parameters of ApiGetResourceDependencyExpression.
type apiGetResourceDependencyExpressionRequest struct {
	HResource mscmrp.HRES_RPC
}

func (*apiGetResourceDependencyExpressionRequest) Opnum() uint16 {
	return clusapi.OpnumApiGetResourceDependencyExpression
}

// apiGetResourceDependencyExpressionResponse carries the [out] parameters and return value of ApiGetResourceDependencyExpression.
type apiGetResourceDependencyExpressionResponse struct {
	LpszDependencyExpression ndr.WSTR
	Rpc_status               ndr.DWORD
	Status                   ndr.DWORD `ndr:"retval"`
}

// ApiGetResourceDependencyExpression calls ApiGetResourceDependencyExpression (opnum 110) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetResourceDependencyExpression(rpc ndr.Invoker, hResource mscmrp.HRES_RPC) (LpszDependencyExpression ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetResourceDependencyExpressionRequest{
		HResource: hResource,
	}
	var resp apiGetResourceDependencyExpressionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetResourceDependencyExpression: %w", err)
		return
	}
	LpszDependencyExpression = resp.LpszDependencyExpression
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetResourceDependencyExpression failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
