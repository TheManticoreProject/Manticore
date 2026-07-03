package functions

import (
	"fmt"

	authzr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraa "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raa"
)

// authzrAccessCheckRequest carries the [in] parameters of AuthzrAccessCheck.
type authzrAccessCheckRequest struct {
	ContextHandle           msraa.AUTHZR_HANDLE
	Flags                   ndr.DWORD
	PRequest                msraa.AUTHZR_ACCESS_REQUEST
	SecurityDescriptorCount ndr.DWORD
	PSecurityDescriptors    []msraa.SR_SD `ndr:"ref,size_is=SecurityDescriptorCount"`
	PReply                  msraa.AUTHZR_ACCESS_REPLY
}

func (*authzrAccessCheckRequest) Opnum() uint16 { return authzr.OpnumAuthzrAccessCheck }

// authzrAccessCheckResponse carries the [out] parameters and return value of AuthzrAccessCheck.
type authzrAccessCheckResponse struct {
	PReply msraa.AUTHZR_ACCESS_REPLY
	Status ndr.DWORD `ndr:"retval"`
}

// AuthzrAccessCheck calls AuthzrAccessCheck (opnum 3) ([MS-RAA] — verify the parameter
// modeling and status handling).
func AuthzrAccessCheck(rpc ndr.Invoker, contextHandle msraa.AUTHZR_HANDLE, flags ndr.DWORD, pRequest msraa.AUTHZR_ACCESS_REQUEST, securityDescriptorCount ndr.DWORD, pSecurityDescriptors []msraa.SR_SD, pReply msraa.AUTHZR_ACCESS_REPLY) (PReply msraa.AUTHZR_ACCESS_REPLY, err error) {
	req := &authzrAccessCheckRequest{
		ContextHandle:           contextHandle,
		Flags:                   flags,
		PRequest:                pRequest,
		SecurityDescriptorCount: securityDescriptorCount,
		PSecurityDescriptors:    pSecurityDescriptors,
		PReply:                  pReply,
	}
	var resp authzrAccessCheckResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("AuthzrAccessCheck: %w", err)
		return
	}
	PReply = resp.PReply
	if uint32(resp.Status) != authzr.StatusSuccess {
		err = fmt.Errorf("AuthzrAccessCheck failed: %s", authzr.StatusString(uint32(resp.Status)))
	}
	return
}
