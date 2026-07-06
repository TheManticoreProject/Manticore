package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncDeletePrinterDriverExRequest carries the [in] parameters of RpcAsyncDeletePrinterDriverEx.
type rpcAsyncDeletePrinterDriverExRequest struct {
	PName        *ndr.WSTR `ndr:"unique"`
	PEnvironment ndr.WSTR
	PDriverName  ndr.WSTR
	DwDeleteFlag ndr.DWORD
	DwVersionNum ndr.DWORD
}

func (*rpcAsyncDeletePrinterDriverExRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeletePrinterDriverEx
}

// rpcAsyncDeletePrinterDriverExResponse carries the [out] parameters and return value of RpcAsyncDeletePrinterDriverEx.
type rpcAsyncDeletePrinterDriverExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeletePrinterDriverEx calls RpcAsyncDeletePrinterDriverEx (opnum 43) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeletePrinterDriverEx(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment ndr.WSTR, pDriverName ndr.WSTR, dwDeleteFlag ndr.DWORD, dwVersionNum ndr.DWORD) (err error) {
	req := &rpcAsyncDeletePrinterDriverExRequest{
		PName:        pName,
		PEnvironment: pEnvironment,
		PDriverName:  pDriverName,
		DwDeleteFlag: dwDeleteFlag,
		DwVersionNum: dwVersionNum,
	}
	var resp rpcAsyncDeletePrinterDriverExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeletePrinterDriverEx: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeletePrinterDriverEx failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
