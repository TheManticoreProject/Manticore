package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncSetJobNamedPropertyRequest carries the [in] parameters of RpcAsyncSetJobNamedProperty.
type rpcAsyncSetJobNamedPropertyRequest struct {
	HPrinter  mspar.PRINTER_HANDLE
	JobId     ndr.DWORD
	PProperty mspar.RPC_PrintNamedProperty
}

func (*rpcAsyncSetJobNamedPropertyRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncSetJobNamedProperty
}

// rpcAsyncSetJobNamedPropertyResponse carries the [out] parameters and return value of RpcAsyncSetJobNamedProperty.
type rpcAsyncSetJobNamedPropertyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncSetJobNamedProperty calls RpcAsyncSetJobNamedProperty (opnum 71) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncSetJobNamedProperty(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, jobId ndr.DWORD, pProperty mspar.RPC_PrintNamedProperty) (err error) {
	req := &rpcAsyncSetJobNamedPropertyRequest{
		HPrinter:  hPrinter,
		JobId:     jobId,
		PProperty: pProperty,
	}
	var resp rpcAsyncSetJobNamedPropertyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncSetJobNamedProperty: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncSetJobNamedProperty failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
