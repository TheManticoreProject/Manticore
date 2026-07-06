package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncDeletePrinterDriverRequest carries the [in] parameters of RpcAsyncDeletePrinterDriver.
type rpcAsyncDeletePrinterDriverRequest struct {
	PName        *ndr.WSTR `ndr:"unique"`
	PEnvironment ndr.WSTR
	PDriverName  ndr.WSTR
}

func (*rpcAsyncDeletePrinterDriverRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeletePrinterDriver
}

// rpcAsyncDeletePrinterDriverResponse carries the [out] parameters and return value of RpcAsyncDeletePrinterDriver.
type rpcAsyncDeletePrinterDriverResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeletePrinterDriver calls RpcAsyncDeletePrinterDriver (opnum 42) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeletePrinterDriver(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment ndr.WSTR, pDriverName ndr.WSTR) (err error) {
	req := &rpcAsyncDeletePrinterDriverRequest{
		PName:        pName,
		PEnvironment: pEnvironment,
		PDriverName:  pDriverName,
	}
	var resp rpcAsyncDeletePrinterDriverResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeletePrinterDriver: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeletePrinterDriver failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
