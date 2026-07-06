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

// authzrFreeContextRequest carries the [in] parameters of AuthzrFreeContext.
type authzrFreeContextRequest struct {
	ContextHandle msraa.AUTHZR_HANDLE
}

func (*authzrFreeContextRequest) Opnum() uint16 { return authzr.OpnumAuthzrFreeContext }

// authzrFreeContextResponse carries the [out] parameters and return value of AuthzrFreeContext.
type authzrFreeContextResponse struct {
	ContextHandle msraa.AUTHZR_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// AuthzrFreeContext calls AuthzrFreeContext (opnum 0) ([MS-RAA] — verify the parameter
// modeling and status handling).
func AuthzrFreeContext(rpc ndr.Invoker, contextHandle msraa.AUTHZR_HANDLE) (ContextHandle msraa.AUTHZR_HANDLE, err error) {
	req := &authzrFreeContextRequest{
		ContextHandle: contextHandle,
	}
	var resp authzrFreeContextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("AuthzrFreeContext: %w", err)
		return
	}
	ContextHandle = resp.ContextHandle
	if uint32(resp.Status) != authzr.StatusSuccess {
		err = fmt.Errorf("AuthzrFreeContext failed: %s", authzr.StatusString(uint32(resp.Status)))
	}
	return
}
