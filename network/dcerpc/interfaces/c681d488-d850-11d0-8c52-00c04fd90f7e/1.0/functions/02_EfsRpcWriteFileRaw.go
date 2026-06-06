package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcWriteFileRawRequest carries the [in] parameters of EfsRpcWriteFileRaw: the
// import context handle and the raw encrypted stream. EfsInPipe is an NDR pipe ([C706]
// 14.7) sent as chunks after the handle.
type efsRpcWriteFileRawRequest struct {
	HContext  structures.PEXIMPORT_CONTEXT_HANDLE
	EfsInPipe structures.EFS_EXIM_PIPE `ndr:"pipe"`
}

func (*efsRpcWriteFileRawRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcWriteFileRaw }

// efsRpcWriteFileRawResponse carries the return value of EfsRpcWriteFileRaw.
type efsRpcWriteFileRawResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcWriteFileRaw calls EfsRpcWriteFileRaw (opnum 2, [MS-EFSR] 3.1.4.2.3). It imports
// a raw encrypted stream into the opened file.
func EfsRpcWriteFileRaw(rpc ndr.Invoker, hContext structures.PEXIMPORT_CONTEXT_HANDLE, efsInPipe structures.EFS_EXIM_PIPE) error {
	req := &efsRpcWriteFileRawRequest{HContext: hContext, EfsInPipe: efsInPipe}
	var resp efsRpcWriteFileRawResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("EfsRpcWriteFileRaw: %w", err)
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		return fmt.Errorf("EfsRpcWriteFileRaw failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
