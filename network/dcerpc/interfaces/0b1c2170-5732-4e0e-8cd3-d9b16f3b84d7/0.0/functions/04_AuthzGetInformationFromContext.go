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

// authzGetInformationFromContextRequest carries the [in] parameters of AuthzGetInformationFromContext.
type authzGetInformationFromContextRequest struct {
	ContextHandle msraa.AUTHZR_HANDLE
	InfoClass     msraa.AUTHZ_CONTEXT_INFORMATION_CLASS `ndr:"enum"`
}

func (*authzGetInformationFromContextRequest) Opnum() uint16 {
	return authzr.OpnumAuthzGetInformationFromContext
}

// authzGetInformationFromContextResponse carries the [out] parameters and return value of AuthzGetInformationFromContext.
type authzGetInformationFromContextResponse struct {
	PpContextInformation *msraa.AUTHZR_CONTEXT_INFORMATION `ndr:"unique"`
	Status               ndr.DWORD                         `ndr:"retval"`
}

// AuthzGetInformationFromContext calls AuthzGetInformationFromContext (opnum 4) ([MS-RAA] — verify the parameter
// modeling and status handling).
func AuthzGetInformationFromContext(rpc ndr.Invoker, contextHandle msraa.AUTHZR_HANDLE, infoClass msraa.AUTHZ_CONTEXT_INFORMATION_CLASS) (PpContextInformation *msraa.AUTHZR_CONTEXT_INFORMATION, err error) {
	req := &authzGetInformationFromContextRequest{
		ContextHandle: contextHandle,
		InfoClass:     infoClass,
	}
	var resp authzGetInformationFromContextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("AuthzGetInformationFromContext: %w", err)
		return
	}
	PpContextInformation = resp.PpContextInformation
	if uint32(resp.Status) != authzr.StatusSuccess {
		err = fmt.Errorf("AuthzGetInformationFromContext failed: %s", authzr.StatusString(uint32(resp.Status)))
	}
	return
}
