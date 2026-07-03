package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcAddPrinterDriverRequest carries the [in] parameters of RpcAddPrinterDriver.
type rpcAddPrinterDriverRequest struct {
	PName            *ndr.WSTR `ndr:"unique"`
	PDriverContainer msrprn.DRIVER_CONTAINER
}

func (*rpcAddPrinterDriverRequest) Opnum() uint16 { return winspool.OpnumRpcAddPrinterDriver }

// rpcAddPrinterDriverResponse carries the [out] parameters and return value of RpcAddPrinterDriver.
type rpcAddPrinterDriverResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAddPrinterDriver calls RpcAddPrinterDriver (opnum 9) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddPrinterDriver(rpc ndr.Invoker, pName *ndr.WSTR, pDriverContainer msrprn.DRIVER_CONTAINER) (err error) {
	req := &rpcAddPrinterDriverRequest{
		PName:            pName,
		PDriverContainer: pDriverContainer,
	}
	var resp rpcAddPrinterDriverResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAddPrinterDriver: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcAddPrinterDriver failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
