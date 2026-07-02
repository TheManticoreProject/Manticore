package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcDeletePrinterDriverRequest carries the [in] parameters of RpcDeletePrinterDriver.
type rpcDeletePrinterDriverRequest struct {
	PName        *ndr.WSTR `ndr:"unique"`
	PEnvironment ndr.WSTR
	PDriverName  ndr.WSTR
}

func (*rpcDeletePrinterDriverRequest) Opnum() uint16 { return winspool.OpnumRpcDeletePrinterDriver }

// rpcDeletePrinterDriverResponse carries the [out] parameters and return value of RpcDeletePrinterDriver.
type rpcDeletePrinterDriverResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeletePrinterDriver calls RpcDeletePrinterDriver (opnum 13) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePrinterDriver(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment ndr.WSTR, pDriverName ndr.WSTR) (err error) {
	req := &rpcDeletePrinterDriverRequest{
		PName:        pName,
		PEnvironment: pEnvironment,
		PDriverName:  pDriverName,
	}
	var resp rpcDeletePrinterDriverResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeletePrinterDriver: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeletePrinterDriver failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
