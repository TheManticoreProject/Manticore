package functions

// IDL source: [MS-RAA] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-raa/0cae6068-686e-4f85-b064-7ba70d47da44
// A fetched copy is kept at ms-raa.idl in the interface directory.

import (
	"fmt"

	authzr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraa "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raa"
)

// authzrModifyClaimsRequest carries the [in] parameters of AuthzrModifyClaims.
type authzrModifyClaimsRequest struct {
	ContextHandle    msraa.AUTHZR_HANDLE
	ClaimClass       msraa.AUTHZ_CONTEXT_INFORMATION_CLASS `ndr:"enum"`
	OperationCount   ndr.DWORD
	PClaimOperations []msraa.AUTHZ_SECURITY_ATTRIBUTE_OPERATION    `ndr:"ref,size_is=OperationCount"`
	PClaims          *msraa.AUTHZR_SECURITY_ATTRIBUTES_INFORMATION `ndr:"unique"`
}

func (*authzrModifyClaimsRequest) Opnum() uint16 { return authzr.OpnumAuthzrModifyClaims }

// authzrModifyClaimsResponse carries the [out] parameters and return value of AuthzrModifyClaims.
type authzrModifyClaimsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// AuthzrModifyClaims calls AuthzrModifyClaims (opnum 5) ([MS-RAA] — verify the parameter
// modeling and status handling).
func AuthzrModifyClaims(rpc ndr.Invoker, contextHandle msraa.AUTHZR_HANDLE, claimClass msraa.AUTHZ_CONTEXT_INFORMATION_CLASS, operationCount ndr.DWORD, pClaimOperations []msraa.AUTHZ_SECURITY_ATTRIBUTE_OPERATION, pClaims *msraa.AUTHZR_SECURITY_ATTRIBUTES_INFORMATION) (err error) {
	req := &authzrModifyClaimsRequest{
		ContextHandle:    contextHandle,
		ClaimClass:       claimClass,
		OperationCount:   operationCount,
		PClaimOperations: pClaimOperations,
		PClaims:          pClaims,
	}
	var resp authzrModifyClaimsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("AuthzrModifyClaims: %w", err)
		return
	}
	if uint32(resp.Status) != authzr.StatusSuccess {
		err = fmt.Errorf("AuthzrModifyClaims failed: %s", authzr.StatusString(uint32(resp.Status)))
	}
	return
}
