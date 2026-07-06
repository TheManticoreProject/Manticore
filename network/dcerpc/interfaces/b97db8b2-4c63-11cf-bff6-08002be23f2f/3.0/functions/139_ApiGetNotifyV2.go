package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetNotifyV2Request carries the [in] parameters of ApiGetNotifyV2.
type apiGetNotifyV2Request struct {
	HNotify mscmrp.HNOTIFY_RPC
}

func (*apiGetNotifyV2Request) Opnum() uint16 { return clusapi.OpnumApiGetNotifyV2 }

// apiGetNotifyV2Response carries the [out] parameters and return value of ApiGetNotifyV2.
type apiGetNotifyV2Response struct {
	Notifications      []*mscmrp.NOTIFICATION_RPC `ndr:"elem=unique,ref,size_is=DwNumNotifications"`
	DwNumNotifications ndr.DWORD
	Status             ndr.DWORD `ndr:"retval"`
}

// ApiGetNotifyV2 calls ApiGetNotifyV2 (opnum 139) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetNotifyV2(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC) (Notifications []*mscmrp.NOTIFICATION_RPC, DwNumNotifications ndr.DWORD, err error) {
	req := &apiGetNotifyV2Request{
		HNotify: hNotify,
	}
	var resp apiGetNotifyV2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetNotifyV2: %w", err)
		return
	}
	Notifications = resp.Notifications
	DwNumNotifications = resp.DwNumNotifications
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetNotifyV2 failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
