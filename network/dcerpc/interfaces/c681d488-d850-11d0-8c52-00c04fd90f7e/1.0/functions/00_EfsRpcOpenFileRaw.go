package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcOpenFileRawRequest carries the [in] parameters of EfsRpcOpenFileRaw.
type efsRpcOpenFileRawRequest struct {
	FileName *ndr.WSTR `ndr:"unique"`
	Flags    int32
}

func (*efsRpcOpenFileRawRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcOpenFileRaw }

// efsRpcOpenFileRawResponse carries the [out] parameters and return value of EfsRpcOpenFileRaw.
type efsRpcOpenFileRawResponse struct {
	HContext structures.PEXIMPORT_CONTEXT_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// EfsRpcOpenFileRaw calls EfsRpcOpenFileRaw (opnum 0) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcOpenFileRaw(rpc ndr.Invoker, fileName *ndr.WSTR, flags int32) (HContext structures.PEXIMPORT_CONTEXT_HANDLE, err error) {
	req := &efsRpcOpenFileRawRequest{
		FileName: fileName,
		Flags:    flags,
	}
	var resp efsRpcOpenFileRawResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcOpenFileRaw: %w", err)
		return
	}
	HContext = resp.HContext
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcOpenFileRaw failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
