package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOpenNetInterfaceRequest carries the [in] parameters of ApiOpenNetInterface.
type apiOpenNetInterfaceRequest struct {
	LpszNetInterfaceName ndr.WSTR
}

func (*apiOpenNetInterfaceRequest) Opnum() uint16 { return clusapi.OpnumApiOpenNetInterface }

// apiOpenNetInterfaceResponse carries the [out] parameters and return value of ApiOpenNetInterface.
type apiOpenNetInterfaceResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HNETINTERFACE_RPC `ndr:"retval"`
}

// ApiOpenNetInterface calls ApiOpenNetInterface (opnum 92) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenNetInterface(rpc ndr.Invoker, lpszNetInterfaceName ndr.WSTR) (Handle mscmrp.HNETINTERFACE_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenNetInterfaceRequest{
		LpszNetInterfaceName: lpszNetInterfaceName,
	}
	var resp apiOpenNetInterfaceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenNetInterface: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenNetInterface failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
