package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiSetGroupSetDependencyExpressionRequest carries the [in] parameters of ApiSetGroupSetDependencyExpression.
type apiSetGroupSetDependencyExpressionRequest struct {
	HGroupSet                mscmrp.HGROUPSET_RPC
	LpszDependencyExpression ndr.WSTR
}

func (*apiSetGroupSetDependencyExpressionRequest) Opnum() uint16 {
	return clusapi.OpnumApiSetGroupSetDependencyExpression
}

// apiSetGroupSetDependencyExpressionResponse carries the [out] parameters and return value of ApiSetGroupSetDependencyExpression.
type apiSetGroupSetDependencyExpressionResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiSetGroupSetDependencyExpression calls ApiSetGroupSetDependencyExpression (opnum 177) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetGroupSetDependencyExpression(rpc ndr.Invoker, hGroupSet mscmrp.HGROUPSET_RPC, lpszDependencyExpression ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiSetGroupSetDependencyExpressionRequest{
		HGroupSet:                hGroupSet,
		LpszDependencyExpression: lpszDependencyExpression,
	}
	var resp apiSetGroupSetDependencyExpressionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetGroupSetDependencyExpression: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetGroupSetDependencyExpression failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
