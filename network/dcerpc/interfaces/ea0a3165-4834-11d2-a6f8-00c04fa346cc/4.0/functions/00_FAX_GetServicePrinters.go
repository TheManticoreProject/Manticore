package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetServicePrintersRequest carries the [in] parameters of FAX_GetServicePrinters.
type fAX_GetServicePrintersRequest struct {
}

func (*fAX_GetServicePrintersRequest) Opnum() uint16 { return fax.OpnumFAX_GetServicePrinters }

// fAX_GetServicePrintersResponse carries the [out] parameters and return value of FAX_GetServicePrinters.
type fAX_GetServicePrintersResponse struct {
	LpBuffer             []byte `ndr:"unique,conformant"`
	LpdwBufferSize       ndr.DWORD
	LpdwPrintersReturned ndr.DWORD
	Status               ndr.DWORD `ndr:"retval"`
}

// FAX_GetServicePrinters calls FAX_GetServicePrinters (opnum 0) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetServicePrinters(rpc ndr.Invoker) (LpBuffer []byte, LpdwBufferSize ndr.DWORD, LpdwPrintersReturned ndr.DWORD, err error) {
	req := &fAX_GetServicePrintersRequest{}
	var resp fAX_GetServicePrintersResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetServicePrinters: %w", err)
		return
	}
	LpBuffer = resp.LpBuffer
	LpdwBufferSize = resp.LpdwBufferSize
	LpdwPrintersReturned = resp.LpdwPrintersReturned
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetServicePrinters failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
