package functions

import (
	"fmt"

	qmcomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76d12b80-3467-11d3-91ff-0090272f9ea3/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// rpc_ACCreateCursorExRequest carries the [in] parameters of rpc_ACCreateCursorEx.
type rpc_ACCreateCursorExRequest struct {
	HQueue msmqmp.RPC_QUEUE_HANDLE
	Pcc    msmqmp.CACCreateRemoteCursor
}

func (*rpc_ACCreateCursorExRequest) Opnum() uint16 { return qmcomm2.Opnumrpc_ACCreateCursorEx }

// rpc_ACCreateCursorExResponse carries the [out] parameters and return value of rpc_ACCreateCursorEx.
type rpc_ACCreateCursorExResponse struct {
	Pcc    msmqmp.CACCreateRemoteCursor
	Status ndr.DWORD `ndr:"retval"`
}

// Rpc_ACCreateCursorEx calls rpc_ACCreateCursorEx (opnum 3) ([MS-MQMP] 3.1.5.5). The
// [in, out] CACCreateRemoteCursor is a plain triple of DWORDs, so this method needs none of
// the CACTransferBuffer double-indirection and is fully implemented.
func Rpc_ACCreateCursorEx(rpc ndr.Invoker, hQueue msmqmp.RPC_QUEUE_HANDLE, pcc msmqmp.CACCreateRemoteCursor) (Pcc msmqmp.CACCreateRemoteCursor, err error) {
	req := &rpc_ACCreateCursorExRequest{
		HQueue: hQueue,
		Pcc:    pcc,
	}
	var resp rpc_ACCreateCursorExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("rpc_ACCreateCursorEx: %w", err)
		return
	}
	Pcc = resp.Pcc
	if uint32(resp.Status) != qmcomm2.StatusSuccess {
		err = fmt.Errorf("rpc_ACCreateCursorEx failed: %s", qmcomm2.StatusString(uint32(resp.Status)))
	}
	return
}
