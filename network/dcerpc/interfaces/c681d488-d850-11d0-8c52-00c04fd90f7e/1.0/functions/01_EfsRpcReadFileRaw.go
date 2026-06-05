package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcReadFileRawRequest carries the [in] parameters of EfsRpcReadFileRaw.
type efsRpcReadFileRawRequest struct {
	HContext structures.PEXIMPORT_CONTEXT_HANDLE
}

func (*efsRpcReadFileRawRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcReadFileRaw }

// efsRpcReadFileRawResponse carries the [out] parameters and return value of EfsRpcReadFileRaw.
type efsRpcReadFileRawResponse struct {
	EfsOutPipe uint8
	Status     ndr.DWORD `ndr:"retval"`
}

// EfsRpcReadFileRaw calls EfsRpcReadFileRaw (opnum 1) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcReadFileRaw(rpc ndr.Invoker, hContext structures.PEXIMPORT_CONTEXT_HANDLE) (EfsOutPipe uint8, err error) {
	req := &efsRpcReadFileRawRequest{
		HContext: hContext,
	}
	var resp efsRpcReadFileRawResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcReadFileRaw: %w", err)
		return
	}
	EfsOutPipe = resp.EfsOutPipe
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcReadFileRaw failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
