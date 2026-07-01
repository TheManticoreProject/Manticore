package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcSendRecvBidiDataRequest carries the [in] parameters of RpcSendRecvBidiData.
type rpcSendRecvBidiDataRequest struct {
	HPrinter structures.PRINTER_HANDLE
	PAction  *ndr.WSTR `ndr:"unique"`
	PReqData structures.RPC_BIDI_REQUEST_CONTAINER
}

func (*rpcSendRecvBidiDataRequest) Opnum() uint16 { return winspool.OpnumRpcSendRecvBidiData }

// rpcSendRecvBidiDataResponse carries the [out] parameters and return value of RpcSendRecvBidiData.
type rpcSendRecvBidiDataResponse struct {
	PpRespData *structures.RPC_BIDI_RESPONSE_CONTAINER `ndr:"unique"`
	Status     ndr.DWORD                               `ndr:"retval"`
}

// RpcSendRecvBidiData calls RpcSendRecvBidiData (opnum 97) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcSendRecvBidiData(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pAction *ndr.WSTR, pReqData structures.RPC_BIDI_REQUEST_CONTAINER) (PpRespData *structures.RPC_BIDI_RESPONSE_CONTAINER, err error) {
	req := &rpcSendRecvBidiDataRequest{
		HPrinter: hPrinter,
		PAction:  pAction,
		PReqData: pReqData,
	}
	var resp rpcSendRecvBidiDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSendRecvBidiData: %w", err)
		return
	}
	PpRespData = resp.PpRespData
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcSendRecvBidiData failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
