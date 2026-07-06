package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRasAdminConnectionNotificationRequest carries the [in] parameters of RRasAdminConnectionNotification.
type rRasAdminConnectionNotificationRequest struct {
	FRegister          ndr.DWORD
	DwClientProcessId  ndr.DWORD
	HEventNotification ndr.DWORD
}

func (*rRasAdminConnectionNotificationRequest) Opnum() uint16 {
	return dimsvc.OpnumRRasAdminConnectionNotification
}

// rRasAdminConnectionNotificationResponse carries the [out] parameters and return value of RRasAdminConnectionNotification.
type rRasAdminConnectionNotificationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRasAdminConnectionNotification calls RRasAdminConnectionNotification (opnum 34) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminConnectionNotification(rpc ndr.Invoker, fRegister ndr.DWORD, dwClientProcessId ndr.DWORD, hEventNotification ndr.DWORD) (err error) {
	req := &rRasAdminConnectionNotificationRequest{
		FRegister:          fRegister,
		DwClientProcessId:  dwClientProcessId,
		HEventNotification: hEventNotification,
	}
	var resp rRasAdminConnectionNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminConnectionNotification: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminConnectionNotification failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
