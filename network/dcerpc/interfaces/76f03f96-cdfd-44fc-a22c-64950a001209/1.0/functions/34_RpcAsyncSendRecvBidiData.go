package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncSendRecvBidiDataRequest carries the [in] parameters of RpcAsyncSendRecvBidiData.
type rpcAsyncSendRecvBidiDataRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	PAction  *ndr.WSTR `ndr:"unique"`
	PReqData mspar.RPC_BIDI_REQUEST_CONTAINER
}

func (*rpcAsyncSendRecvBidiDataRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncSendRecvBidiData
}

// rpcAsyncSendRecvBidiDataResponse carries the [out] parameters and return value of RpcAsyncSendRecvBidiData.
type rpcAsyncSendRecvBidiDataResponse struct {
	PpRespData mspar.RPC_BIDI_RESPONSE_CONTAINER
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcAsyncSendRecvBidiData calls RpcAsyncSendRecvBidiData (opnum 34) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncSendRecvBidiData(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pAction *ndr.WSTR, pReqData mspar.RPC_BIDI_REQUEST_CONTAINER) (PpRespData mspar.RPC_BIDI_RESPONSE_CONTAINER, err error) {
	req := &rpcAsyncSendRecvBidiDataRequest{
		HPrinter: hPrinter,
		PAction:  pAction,
		PReqData: pReqData,
	}
	var resp rpcAsyncSendRecvBidiDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncSendRecvBidiData: %w", err)
		return
	}
	PpRespData = resp.PpRespData
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncSendRecvBidiData failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
