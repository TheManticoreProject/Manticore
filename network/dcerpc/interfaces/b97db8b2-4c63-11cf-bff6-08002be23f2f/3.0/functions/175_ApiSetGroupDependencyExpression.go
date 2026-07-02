package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiSetGroupDependencyExpressionRequest carries the [in] parameters of ApiSetGroupDependencyExpression.
type apiSetGroupDependencyExpressionRequest struct {
	HGroup                   mscmrp.HGROUP_RPC
	LpszDependencyExpression ndr.WSTR
}

func (*apiSetGroupDependencyExpressionRequest) Opnum() uint16 {
	return clusapi.OpnumApiSetGroupDependencyExpression
}

// apiSetGroupDependencyExpressionResponse carries the [out] parameters and return value of ApiSetGroupDependencyExpression.
type apiSetGroupDependencyExpressionResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiSetGroupDependencyExpression calls ApiSetGroupDependencyExpression (opnum 175) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetGroupDependencyExpression(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, lpszDependencyExpression ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiSetGroupDependencyExpressionRequest{
		HGroup:                   hGroup,
		LpszDependencyExpression: lpszDependencyExpression,
	}
	var resp apiSetGroupDependencyExpressionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetGroupDependencyExpression: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetGroupDependencyExpression failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
