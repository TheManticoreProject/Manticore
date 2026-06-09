package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcEncryptFileSrvRequest carries the [in] parameters of EfsRpcEncryptFileSrv.
type efsRpcEncryptFileSrvRequest struct {
	FileName ndr.WSTR
}

func (*efsRpcEncryptFileSrvRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcEncryptFileSrv }

// efsRpcEncryptFileSrvResponse carries the [out] parameters and return value of EfsRpcEncryptFileSrv.
type efsRpcEncryptFileSrvResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcEncryptFileSrv calls EfsRpcEncryptFileSrv (opnum 4) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcEncryptFileSrv(rpc ndr.Invoker, fileName ndr.WSTR) (err error) {
	req := &efsRpcEncryptFileSrvRequest{
		FileName: fileName,
	}
	var resp efsRpcEncryptFileSrvResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcEncryptFileSrv: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcEncryptFileSrv failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
