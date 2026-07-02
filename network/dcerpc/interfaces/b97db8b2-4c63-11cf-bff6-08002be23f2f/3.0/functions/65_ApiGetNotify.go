package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetNotifyRequest carries the [in] parameters of ApiGetNotify.
type apiGetNotifyRequest struct {
	HNotify mscmrp.HNOTIFY_RPC
}

func (*apiGetNotifyRequest) Opnum() uint16 { return clusapi.OpnumApiGetNotify }

// apiGetNotifyResponse carries the [out] parameters and return value of ApiGetNotify.
type apiGetNotifyResponse struct {
	DwNotifyKey     ndr.DWORD
	DwFilter        ndr.DWORD
	DwStateSequence ndr.DWORD
	Name            ndr.WSTR
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiGetNotify calls ApiGetNotify (opnum 65) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetNotify(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC) (DwNotifyKey ndr.DWORD, DwFilter ndr.DWORD, DwStateSequence ndr.DWORD, Name ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetNotifyRequest{
		HNotify: hNotify,
	}
	var resp apiGetNotifyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetNotify: %w", err)
		return
	}
	DwNotifyKey = resp.DwNotifyKey
	DwFilter = resp.DwFilter
	DwStateSequence = resp.DwStateSequence
	Name = resp.Name
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetNotify failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
