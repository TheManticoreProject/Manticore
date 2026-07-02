package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_SetDeviceOrderInGroupRequest carries the [in] parameters of FAX_SetDeviceOrderInGroup.
type fAX_SetDeviceOrderInGroupRequest struct {
	LpwstrGroupName ndr.WSTR
	DwDeviceId      ndr.DWORD
	DwNewOrder      ndr.DWORD
}

func (*fAX_SetDeviceOrderInGroupRequest) Opnum() uint16 { return fax.OpnumFAX_SetDeviceOrderInGroup }

// fAX_SetDeviceOrderInGroupResponse carries the [out] parameters and return value of FAX_SetDeviceOrderInGroup.
type fAX_SetDeviceOrderInGroupResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetDeviceOrderInGroup calls FAX_SetDeviceOrderInGroup (opnum 55) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetDeviceOrderInGroup(rpc ndr.Invoker, lpwstrGroupName ndr.WSTR, dwDeviceId ndr.DWORD, dwNewOrder ndr.DWORD) (err error) {
	req := &fAX_SetDeviceOrderInGroupRequest{
		LpwstrGroupName: lpwstrGroupName,
		DwDeviceId:      dwDeviceId,
		DwNewOrder:      dwNewOrder,
	}
	var resp fAX_SetDeviceOrderInGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetDeviceOrderInGroup: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetDeviceOrderInGroup failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
