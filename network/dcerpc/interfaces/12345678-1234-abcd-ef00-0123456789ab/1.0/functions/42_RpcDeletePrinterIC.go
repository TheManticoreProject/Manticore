package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcDeletePrinterICRequest carries the [in] parameters of RpcDeletePrinterIC.
type rpcDeletePrinterICRequest struct {
	PhPrinterIC msrprn.GDI_HANDLE
}

func (*rpcDeletePrinterICRequest) Opnum() uint16 { return winspool.OpnumRpcDeletePrinterIC }

// rpcDeletePrinterICResponse carries the [out] parameters and return value of RpcDeletePrinterIC.
type rpcDeletePrinterICResponse struct {
	PhPrinterIC msrprn.GDI_HANDLE
	Status      ndr.DWORD `ndr:"retval"`
}

// RpcDeletePrinterIC calls RpcDeletePrinterIC (opnum 42) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePrinterIC(rpc ndr.Invoker, phPrinterIC msrprn.GDI_HANDLE) (PhPrinterIC msrprn.GDI_HANDLE, err error) {
	req := &rpcDeletePrinterICRequest{
		PhPrinterIC: phPrinterIC,
	}
	var resp rpcDeletePrinterICResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeletePrinterIC: %w", err)
		return
	}
	PhPrinterIC = resp.PhPrinterIC
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeletePrinterIC failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
