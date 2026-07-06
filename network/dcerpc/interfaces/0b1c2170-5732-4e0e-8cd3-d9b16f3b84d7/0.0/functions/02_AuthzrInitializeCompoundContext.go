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

// authzrInitializeCompoundContextRequest carries the [in] parameters of AuthzrInitializeCompoundContext.
type authzrInitializeCompoundContextRequest struct {
	UserContextHandle   msraa.AUTHZR_HANDLE
	DeviceContextHandle msraa.AUTHZR_HANDLE
}

func (*authzrInitializeCompoundContextRequest) Opnum() uint16 {
	return authzr.OpnumAuthzrInitializeCompoundContext
}

// authzrInitializeCompoundContextResponse carries the [out] parameters and return value of AuthzrInitializeCompoundContext.
type authzrInitializeCompoundContextResponse struct {
	CompoundContextHandle msraa.AUTHZR_HANDLE
	Status                ndr.DWORD `ndr:"retval"`
}

// AuthzrInitializeCompoundContext calls AuthzrInitializeCompoundContext (opnum 2) ([MS-RAA] — verify the parameter
// modeling and status handling).
func AuthzrInitializeCompoundContext(rpc ndr.Invoker, userContextHandle msraa.AUTHZR_HANDLE, deviceContextHandle msraa.AUTHZR_HANDLE) (CompoundContextHandle msraa.AUTHZR_HANDLE, err error) {
	req := &authzrInitializeCompoundContextRequest{
		UserContextHandle:   userContextHandle,
		DeviceContextHandle: deviceContextHandle,
	}
	var resp authzrInitializeCompoundContextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("AuthzrInitializeCompoundContext: %w", err)
		return
	}
	CompoundContextHandle = resp.CompoundContextHandle
	if uint32(resp.Status) != authzr.StatusSuccess {
		err = fmt.Errorf("AuthzrInitializeCompoundContext failed: %s", authzr.StatusString(uint32(resp.Status)))
	}
	return
}
