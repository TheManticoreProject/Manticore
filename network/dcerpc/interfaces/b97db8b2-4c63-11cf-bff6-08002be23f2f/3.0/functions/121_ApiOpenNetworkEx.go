package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOpenNetworkExRequest carries the [in] parameters of ApiOpenNetworkEx.
type apiOpenNetworkExRequest struct {
	LpszNetworkName ndr.WSTR
	DwDesiredAccess ndr.DWORD
}

func (*apiOpenNetworkExRequest) Opnum() uint16 { return clusapi.OpnumApiOpenNetworkEx }

// apiOpenNetworkExResponse carries the [out] parameters and return value of ApiOpenNetworkEx.
type apiOpenNetworkExResponse struct {
	LpdwGrantedAccess ndr.DWORD
	Status            ndr.DWORD
	Rpc_status        ndr.DWORD
	Handle            mscmrp.HNETWORK_RPC `ndr:"retval"`
}

// ApiOpenNetworkEx calls ApiOpenNetworkEx (opnum 121) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenNetworkEx(rpc ndr.Invoker, lpszNetworkName ndr.WSTR, dwDesiredAccess ndr.DWORD) (Handle mscmrp.HNETWORK_RPC, LpdwGrantedAccess ndr.DWORD, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenNetworkExRequest{
		LpszNetworkName: lpszNetworkName,
		DwDesiredAccess: dwDesiredAccess,
	}
	var resp apiOpenNetworkExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenNetworkEx: %w", err)
		return
	}
	Handle = resp.Handle
	LpdwGrantedAccess = resp.LpdwGrantedAccess
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenNetworkEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
