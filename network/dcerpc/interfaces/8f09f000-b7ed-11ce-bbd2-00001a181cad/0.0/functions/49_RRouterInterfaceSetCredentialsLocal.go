package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRouterInterfaceSetCredentialsLocalRequest carries the [in] parameters of RRouterInterfaceSetCredentialsLocal.
type rRouterInterfaceSetCredentialsLocalRequest struct {
	LpwsInterfaceName ndr.WSTR
	LpwsUserName      ndr.WSTR
	LpwsDomainName    ndr.WSTR
	LpwsPassword      ndr.WSTR
}

func (*rRouterInterfaceSetCredentialsLocalRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceSetCredentialsLocal
}

// rRouterInterfaceSetCredentialsLocalResponse carries the [out] parameters and return value of RRouterInterfaceSetCredentialsLocal.
type rRouterInterfaceSetCredentialsLocalResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceSetCredentialsLocal calls RRouterInterfaceSetCredentialsLocal (opnum 49) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceSetCredentialsLocal(rpc ndr.Invoker, lpwsInterfaceName ndr.WSTR, lpwsUserName ndr.WSTR, lpwsDomainName ndr.WSTR, lpwsPassword ndr.WSTR) (err error) {
	req := &rRouterInterfaceSetCredentialsLocalRequest{
		LpwsInterfaceName: lpwsInterfaceName,
		LpwsUserName:      lpwsUserName,
		LpwsDomainName:    lpwsDomainName,
		LpwsPassword:      lpwsPassword,
	}
	var resp rRouterInterfaceSetCredentialsLocalResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceSetCredentialsLocal: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceSetCredentialsLocal failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
