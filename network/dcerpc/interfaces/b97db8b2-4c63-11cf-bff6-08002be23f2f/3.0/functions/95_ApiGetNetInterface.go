package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// apiGetNetInterfaceRequest carries the [in] parameters of ApiGetNetInterface.
type apiGetNetInterfaceRequest struct {
	LpszNodeName    ndr.WSTR
	LpszNetworkName ndr.WSTR
}

func (*apiGetNetInterfaceRequest) Opnum() uint16 { return clusapi.OpnumApiGetNetInterface }

// apiGetNetInterfaceResponse carries the [out] parameters and return value of ApiGetNetInterface.
type apiGetNetInterfaceResponse struct {
	LppszInterfaceName ndr.WSTR
	Rpc_status         ndr.DWORD
	Status             ndr.DWORD `ndr:"retval"`
}

// ApiGetNetInterface calls ApiGetNetInterface (opnum 95) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetNetInterface(rpc ndr.Invoker, lpszNodeName ndr.WSTR, lpszNetworkName ndr.WSTR) (LppszInterfaceName ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetNetInterfaceRequest{
		LpszNodeName:    lpszNodeName,
		LpszNetworkName: lpszNetworkName,
	}
	var resp apiGetNetInterfaceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetNetInterface: %w", err)
		return
	}
	LppszInterfaceName = resp.LppszInterfaceName
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetNetInterface failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
