package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiSetNetworkNameRequest carries the [in] parameters of ApiSetNetworkName.
type apiSetNetworkNameRequest struct {
	HNetwork        mscmrp.HNETWORK_RPC
	LpszNetworkName ndr.WSTR
}

func (*apiSetNetworkNameRequest) Opnum() uint16 { return clusapi.OpnumApiSetNetworkName }

// apiSetNetworkNameResponse carries the [out] parameters and return value of ApiSetNetworkName.
type apiSetNetworkNameResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiSetNetworkName calls ApiSetNetworkName (opnum 84) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetNetworkName(rpc ndr.Invoker, hNetwork mscmrp.HNETWORK_RPC, lpszNetworkName ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiSetNetworkNameRequest{
		HNetwork:        hNetwork,
		LpszNetworkName: lpszNetworkName,
	}
	var resp apiSetNetworkNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetNetworkName: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetNetworkName failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
