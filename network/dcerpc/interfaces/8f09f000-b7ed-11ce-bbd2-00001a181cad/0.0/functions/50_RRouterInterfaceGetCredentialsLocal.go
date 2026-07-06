package functions

// IDL source: [MS-RRASM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrasm/8e6e89fb-9c80-4a9a-a222-d7d8948244bb
// A fetched copy is kept at ms-rrasm.idl in the interface directory.

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRouterInterfaceGetCredentialsLocalRequest carries the [in] parameters of RRouterInterfaceGetCredentialsLocal.
type rRouterInterfaceGetCredentialsLocalRequest struct {
	LpwsInterfaceName ndr.WSTR
}

func (*rRouterInterfaceGetCredentialsLocalRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceGetCredentialsLocal
}

// rRouterInterfaceGetCredentialsLocalResponse carries the [out] parameters and return value of RRouterInterfaceGetCredentialsLocal.
type rRouterInterfaceGetCredentialsLocalResponse struct {
	LpwsUserName   ndr.WSTR
	LpwsDomainName ndr.WSTR
	LpwsPassword   ndr.WSTR
	Status         ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceGetCredentialsLocal calls RRouterInterfaceGetCredentialsLocal (opnum 50) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceGetCredentialsLocal(rpc ndr.Invoker, lpwsInterfaceName ndr.WSTR) (LpwsUserName ndr.WSTR, LpwsDomainName ndr.WSTR, LpwsPassword ndr.WSTR, err error) {
	req := &rRouterInterfaceGetCredentialsLocalRequest{
		LpwsInterfaceName: lpwsInterfaceName,
	}
	var resp rRouterInterfaceGetCredentialsLocalResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceGetCredentialsLocal: %w", err)
		return
	}
	LpwsUserName = resp.LpwsUserName
	LpwsDomainName = resp.LpwsDomainName
	LpwsPassword = resp.LpwsPassword
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceGetCredentialsLocal failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
