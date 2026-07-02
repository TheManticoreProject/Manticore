package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_GetDeviceStatusRequest carries the [in] parameters of FAX_GetDeviceStatus.
type fAX_GetDeviceStatusRequest struct {
	FaxPortHandle msfax.RPC_FAX_PORT_HANDLE
}

func (*fAX_GetDeviceStatusRequest) Opnum() uint16 { return fax.OpnumFAX_GetDeviceStatus }

// fAX_GetDeviceStatusResponse carries the [out] parameters and return value of FAX_GetDeviceStatus.
type fAX_GetDeviceStatusResponse struct {
	StatusBuffer []byte `ndr:"unique,conformant"`
	BufferSize   ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// FAX_GetDeviceStatus calls FAX_GetDeviceStatus (opnum 8) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetDeviceStatus(rpc ndr.Invoker, faxPortHandle msfax.RPC_FAX_PORT_HANDLE) (StatusBuffer []byte, BufferSize ndr.DWORD, err error) {
	req := &fAX_GetDeviceStatusRequest{
		FaxPortHandle: faxPortHandle,
	}
	var resp fAX_GetDeviceStatusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetDeviceStatus: %w", err)
		return
	}
	StatusBuffer = resp.StatusBuffer
	BufferSize = resp.BufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetDeviceStatus failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
