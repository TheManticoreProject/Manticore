package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// apiDeleteResourceTypeRequest carries the [in] parameters of ApiDeleteResourceType.
type apiDeleteResourceTypeRequest struct {
	LpszTypeName ndr.WSTR
}

func (*apiDeleteResourceTypeRequest) Opnum() uint16 { return clusapi.OpnumApiDeleteResourceType }

// apiDeleteResourceTypeResponse carries the [out] parameters and return value of ApiDeleteResourceType.
type apiDeleteResourceTypeResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiDeleteResourceType calls ApiDeleteResourceType (opnum 27) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiDeleteResourceType(rpc ndr.Invoker, lpszTypeName ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiDeleteResourceTypeRequest{
		LpszTypeName: lpszTypeName,
	}
	var resp apiDeleteResourceTypeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiDeleteResourceType: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiDeleteResourceType failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
