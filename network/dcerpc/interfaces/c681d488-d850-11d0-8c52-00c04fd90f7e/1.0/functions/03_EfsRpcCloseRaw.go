package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcCloseRawRequest carries the [in] parameters of EfsRpcCloseRaw.
type efsRpcCloseRawRequest struct {
	HContext structures.PEXIMPORT_CONTEXT_HANDLE
}

func (*efsRpcCloseRawRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcCloseRaw }

// efsRpcCloseRawResponse carries the [out] parameters and return value of EfsRpcCloseRaw.
type efsRpcCloseRawResponse struct {
	HContext structures.PEXIMPORT_CONTEXT_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// EfsRpcCloseRaw calls EfsRpcCloseRaw (opnum 3) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcCloseRaw(rpc ndr.Invoker, hContext structures.PEXIMPORT_CONTEXT_HANDLE) (HContext structures.PEXIMPORT_CONTEXT_HANDLE, err error) {
	req := &efsRpcCloseRawRequest{
		HContext: hContext,
	}
	var resp efsRpcCloseRawResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcCloseRaw: %w", err)
		return
	}
	HContext = resp.HContext
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcCloseRaw failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
