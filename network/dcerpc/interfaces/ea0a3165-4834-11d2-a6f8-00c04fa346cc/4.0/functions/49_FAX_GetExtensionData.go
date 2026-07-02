package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetExtensionDataRequest carries the [in] parameters of FAX_GetExtensionData.
type fAX_GetExtensionDataRequest struct {
	DwDeviceId      ndr.DWORD
	LpcwstrNameGUID ndr.WSTR
}

func (*fAX_GetExtensionDataRequest) Opnum() uint16 { return fax.OpnumFAX_GetExtensionData }

// fAX_GetExtensionDataResponse carries the [out] parameters and return value of FAX_GetExtensionData.
type fAX_GetExtensionDataResponse struct {
	PpData       []byte `ndr:"unique,conformant"`
	LpdwDataSize ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// FAX_GetExtensionData calls FAX_GetExtensionData (opnum 49) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetExtensionData(rpc ndr.Invoker, dwDeviceId ndr.DWORD, lpcwstrNameGUID ndr.WSTR) (PpData []byte, LpdwDataSize ndr.DWORD, err error) {
	req := &fAX_GetExtensionDataRequest{
		DwDeviceId:      dwDeviceId,
		LpcwstrNameGUID: lpcwstrNameGUID,
	}
	var resp fAX_GetExtensionDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetExtensionData: %w", err)
		return
	}
	PpData = resp.PpData
	LpdwDataSize = resp.LpdwDataSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetExtensionData failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
