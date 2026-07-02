package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiSetResourceDependencyExpressionRequest carries the [in] parameters of ApiSetResourceDependencyExpression.
type apiSetResourceDependencyExpressionRequest struct {
	HResource                mscmrp.HRES_RPC
	LpszDependencyExpression *ndr.WSTR `ndr:"unique"`
}

func (*apiSetResourceDependencyExpressionRequest) Opnum() uint16 {
	return clusapi.OpnumApiSetResourceDependencyExpression
}

// apiSetResourceDependencyExpressionResponse carries the [out] parameters and return value of ApiSetResourceDependencyExpression.
type apiSetResourceDependencyExpressionResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiSetResourceDependencyExpression calls ApiSetResourceDependencyExpression (opnum 109) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetResourceDependencyExpression(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, lpszDependencyExpression *ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiSetResourceDependencyExpressionRequest{
		HResource:                hResource,
		LpszDependencyExpression: lpszDependencyExpression,
	}
	var resp apiSetResourceDependencyExpressionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetResourceDependencyExpression: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetResourceDependencyExpression failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
