package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_SetExtensionDataRequest carries the [in] parameters of FAX_SetExtensionData.
type fAX_SetExtensionDataRequest struct {
	LpcwstrComputerName ndr.WSTR
	DwDeviceId          ndr.DWORD
	LpcwstrNameGUID     ndr.WSTR
	PData               []uint8 `ndr:"ref,size_is=DwDataSize"`
	DwDataSize          ndr.DWORD
}

func (*fAX_SetExtensionDataRequest) Opnum() uint16 { return fax.OpnumFAX_SetExtensionData }

// fAX_SetExtensionDataResponse carries the [out] parameters and return value of FAX_SetExtensionData.
type fAX_SetExtensionDataResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetExtensionData calls FAX_SetExtensionData (opnum 50) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetExtensionData(rpc ndr.Invoker, lpcwstrComputerName ndr.WSTR, dwDeviceId ndr.DWORD, lpcwstrNameGUID ndr.WSTR, pData []uint8, dwDataSize ndr.DWORD) (err error) {
	req := &fAX_SetExtensionDataRequest{
		LpcwstrComputerName: lpcwstrComputerName,
		DwDeviceId:          dwDeviceId,
		LpcwstrNameGUID:     lpcwstrNameGUID,
		PData:               pData,
		DwDataSize:          dwDataSize,
	}
	var resp fAX_SetExtensionDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetExtensionData: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetExtensionData failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
