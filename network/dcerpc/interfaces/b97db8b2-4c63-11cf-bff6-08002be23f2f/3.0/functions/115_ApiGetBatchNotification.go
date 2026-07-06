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

// apiGetBatchNotificationRequest carries the [in] parameters of ApiGetBatchNotification.
type apiGetBatchNotificationRequest struct {
	HBatchNotify mscmrp.HBATCH_PORT_RPC
}

func (*apiGetBatchNotificationRequest) Opnum() uint16 { return clusapi.OpnumApiGetBatchNotification }

// apiGetBatchNotificationResponse carries the [out] parameters and return value of ApiGetBatchNotification.
type apiGetBatchNotificationResponse struct {
	CbData ndr.DWORD
	LpData []uint8   `ndr:"ref,size_is=CbData"`
	Status ndr.DWORD `ndr:"retval"`
}

// ApiGetBatchNotification calls ApiGetBatchNotification (opnum 115) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetBatchNotification(rpc ndr.Invoker, hBatchNotify mscmrp.HBATCH_PORT_RPC) (CbData ndr.DWORD, LpData []uint8, err error) {
	req := &apiGetBatchNotificationRequest{
		HBatchNotify: hBatchNotify,
	}
	var resp apiGetBatchNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetBatchNotification: %w", err)
		return
	}
	CbData = resp.CbData
	LpData = resp.LpData
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetBatchNotification failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
