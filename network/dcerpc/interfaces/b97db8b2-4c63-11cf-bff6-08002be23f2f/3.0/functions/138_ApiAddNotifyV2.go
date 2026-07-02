package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiAddNotifyV2Request carries the [in] parameters of ApiAddNotifyV2.
type apiAddNotifyV2Request struct {
	HNotify            mscmrp.HNOTIFY_RPC
	HObject            mscmrp.HGENERIC_RPC
	Filter             mscmrp.NOTIFY_FILTER_AND_TYPE_RPC
	DwNotifyKey        ndr.DWORD
	DwVersion          ndr.DWORD
	IsTargetedAtObject ndr.BOOL
}

func (*apiAddNotifyV2Request) Opnum() uint16 { return clusapi.OpnumApiAddNotifyV2 }

// apiAddNotifyV2Response carries the [out] parameters and return value of ApiAddNotifyV2.
type apiAddNotifyV2Response struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiAddNotifyV2 calls ApiAddNotifyV2 (opnum 138) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddNotifyV2(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hObject mscmrp.HGENERIC_RPC, filter mscmrp.NOTIFY_FILTER_AND_TYPE_RPC, dwNotifyKey ndr.DWORD, dwVersion ndr.DWORD, isTargetedAtObject ndr.BOOL) (Rpc_status ndr.DWORD, err error) {
	req := &apiAddNotifyV2Request{
		HNotify:            hNotify,
		HObject:            hObject,
		Filter:             filter,
		DwNotifyKey:        dwNotifyKey,
		DwVersion:          dwVersion,
		IsTargetedAtObject: isTargetedAtObject,
	}
	var resp apiAddNotifyV2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddNotifyV2: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddNotifyV2 failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
