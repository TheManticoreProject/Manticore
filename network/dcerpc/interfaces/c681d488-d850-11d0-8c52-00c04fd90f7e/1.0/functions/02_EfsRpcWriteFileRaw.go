package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcWriteFileRawRequest carries the [in] parameters of EfsRpcWriteFileRaw.
type efsRpcWriteFileRawRequest struct {
	HContext  structures.PEXIMPORT_CONTEXT_HANDLE
	EfsInPipe uint8
}

func (*efsRpcWriteFileRawRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcWriteFileRaw }

// efsRpcWriteFileRawResponse carries the [out] parameters and return value of EfsRpcWriteFileRaw.
type efsRpcWriteFileRawResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcWriteFileRaw calls EfsRpcWriteFileRaw (opnum 2) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcWriteFileRaw(rpc ndr.Invoker, hContext structures.PEXIMPORT_CONTEXT_HANDLE, efsInPipe uint8) (err error) {
	req := &efsRpcWriteFileRawRequest{
		HContext:  hContext,
		EfsInPipe: efsInPipe,
	}
	var resp efsRpcWriteFileRawResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcWriteFileRaw: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcWriteFileRaw failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
