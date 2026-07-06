package functions

import (
	"fmt"

	rasrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/20610036-fa22-11cf-9823-00a0c911e5df/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rasRpcSubmitRequestRequest carries the [in] parameters of RasRpcSubmitRequest.
type rasRpcSubmitRequestRequest struct {
	PReqBuffer  []uint8 `ndr:"ref,size_is=DwcbBufSize"`
	DwcbBufSize ndr.DWORD
}

func (*rasRpcSubmitRequestRequest) Opnum() uint16 { return rasrpc.OpnumRasRpcSubmitRequest }

// rasRpcSubmitRequestResponse carries the [out] parameters and return value of RasRpcSubmitRequest.
type rasRpcSubmitRequestResponse struct {
	PReqBuffer []uint8   `ndr:"ref,size_is=DwcbBufSize"`
	Status     ndr.DWORD `ndr:"retval"`
}

// RasRpcSubmitRequest calls RasRpcSubmitRequest (opnum 12) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RasRpcSubmitRequest(rpc ndr.Invoker, pReqBuffer []uint8, dwcbBufSize ndr.DWORD) (PReqBuffer []uint8, err error) {
	req := &rasRpcSubmitRequestRequest{
		PReqBuffer:  pReqBuffer,
		DwcbBufSize: dwcbBufSize,
	}
	var resp rasRpcSubmitRequestResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RasRpcSubmitRequest: %w", err)
		return
	}
	PReqBuffer = resp.PReqBuffer
	if uint32(resp.Status) != rasrpc.StatusSuccess {
		err = fmt.Errorf("RasRpcSubmitRequest failed: %s", rasrpc.StatusString(uint32(resp.Status)))
	}
	return
}
