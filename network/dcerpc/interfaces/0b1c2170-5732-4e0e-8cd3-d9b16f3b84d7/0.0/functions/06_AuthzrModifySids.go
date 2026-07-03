package functions

import (
	"fmt"

	authzr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraa "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raa"
)

// authzrModifySidsRequest carries the [in] parameters of AuthzrModifySids.
type authzrModifySidsRequest struct {
	ContextHandle  msraa.AUTHZR_HANDLE
	SidClass       msraa.AUTHZ_CONTEXT_INFORMATION_CLASS `ndr:"enum"`
	OperationCount ndr.DWORD
	PSidOperations []msraa.AUTHZ_SID_OPERATION `ndr:"ref,size_is=OperationCount"`
	PSids          *msraa.AUTHZR_TOKEN_GROUPS  `ndr:"unique"`
}

func (*authzrModifySidsRequest) Opnum() uint16 { return authzr.OpnumAuthzrModifySids }

// authzrModifySidsResponse carries the [out] parameters and return value of AuthzrModifySids.
type authzrModifySidsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// AuthzrModifySids calls AuthzrModifySids (opnum 6) ([MS-RAA] — verify the parameter
// modeling and status handling).
func AuthzrModifySids(rpc ndr.Invoker, contextHandle msraa.AUTHZR_HANDLE, sidClass msraa.AUTHZ_CONTEXT_INFORMATION_CLASS, operationCount ndr.DWORD, pSidOperations []msraa.AUTHZ_SID_OPERATION, pSids *msraa.AUTHZR_TOKEN_GROUPS) (err error) {
	req := &authzrModifySidsRequest{
		ContextHandle:  contextHandle,
		SidClass:       sidClass,
		OperationCount: operationCount,
		PSidOperations: pSidOperations,
		PSids:          pSids,
	}
	var resp authzrModifySidsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("AuthzrModifySids: %w", err)
		return
	}
	if uint32(resp.Status) != authzr.StatusSuccess {
		err = fmt.Errorf("AuthzrModifySids failed: %s", authzr.StatusString(uint32(resp.Status)))
	}
	return
}
