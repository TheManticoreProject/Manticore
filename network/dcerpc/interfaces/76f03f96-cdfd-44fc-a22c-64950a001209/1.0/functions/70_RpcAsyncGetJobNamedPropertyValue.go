package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncGetJobNamedPropertyValueRequest carries the [in] parameters of RpcAsyncGetJobNamedPropertyValue.
type rpcAsyncGetJobNamedPropertyValueRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	JobId    ndr.DWORD
	PszName  ndr.WSTR
}

func (*rpcAsyncGetJobNamedPropertyValueRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncGetJobNamedPropertyValue
}

// rpcAsyncGetJobNamedPropertyValueResponse carries the [out] parameters and return value of RpcAsyncGetJobNamedPropertyValue.
type rpcAsyncGetJobNamedPropertyValueResponse struct {
	PValue mspar.RPC_PrintPropertyValue
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetJobNamedPropertyValue calls RpcAsyncGetJobNamedPropertyValue (opnum 70) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetJobNamedPropertyValue(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, jobId ndr.DWORD, pszName ndr.WSTR) (PValue mspar.RPC_PrintPropertyValue, err error) {
	req := &rpcAsyncGetJobNamedPropertyValueRequest{
		HPrinter: hPrinter,
		JobId:    jobId,
		PszName:  pszName,
	}
	var resp rpcAsyncGetJobNamedPropertyValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetJobNamedPropertyValue: %w", err)
		return
	}
	PValue = resp.PValue
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetJobNamedPropertyValue failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
