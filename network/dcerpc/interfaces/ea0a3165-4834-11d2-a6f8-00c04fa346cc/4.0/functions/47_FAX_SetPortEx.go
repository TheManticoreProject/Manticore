package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_SetPortExRequest carries the [in] parameters of FAX_SetPortEx.
type fAX_SetPortExRequest struct {
	DwDeviceId ndr.DWORD
	PPortInfo  msfax.FAX_PORT_INFO_EXW
}

func (*fAX_SetPortExRequest) Opnum() uint16 { return fax.OpnumFAX_SetPortEx }

// fAX_SetPortExResponse carries the [out] parameters and return value of FAX_SetPortEx.
type fAX_SetPortExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetPortEx calls FAX_SetPortEx (opnum 47) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetPortEx(rpc ndr.Invoker, dwDeviceId ndr.DWORD, pPortInfo msfax.FAX_PORT_INFO_EXW) (err error) {
	req := &fAX_SetPortExRequest{
		DwDeviceId: dwDeviceId,
		PPortInfo:  pPortInfo,
	}
	var resp fAX_SetPortExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetPortEx: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetPortEx failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
