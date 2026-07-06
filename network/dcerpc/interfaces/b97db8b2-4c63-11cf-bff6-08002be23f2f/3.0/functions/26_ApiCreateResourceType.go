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

// apiCreateResourceTypeRequest carries the [in] parameters of ApiCreateResourceType.
type apiCreateResourceTypeRequest struct {
	LpszTypeName    ndr.WSTR
	LpszDisplayName ndr.WSTR
	LpszDllName     ndr.WSTR
	DwLooksAlive    ndr.DWORD
	DwIsAlive       ndr.DWORD
}

func (*apiCreateResourceTypeRequest) Opnum() uint16 { return clusapi.OpnumApiCreateResourceType }

// apiCreateResourceTypeResponse carries the [out] parameters and return value of ApiCreateResourceType.
type apiCreateResourceTypeResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiCreateResourceType calls ApiCreateResourceType (opnum 26) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateResourceType(rpc ndr.Invoker, lpszTypeName ndr.WSTR, lpszDisplayName ndr.WSTR, lpszDllName ndr.WSTR, dwLooksAlive ndr.DWORD, dwIsAlive ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiCreateResourceTypeRequest{
		LpszTypeName:    lpszTypeName,
		LpszDisplayName: lpszDisplayName,
		LpszDllName:     lpszDllName,
		DwLooksAlive:    dwLooksAlive,
		DwIsAlive:       dwIsAlive,
	}
	var resp apiCreateResourceTypeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateResourceType: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateResourceType failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
