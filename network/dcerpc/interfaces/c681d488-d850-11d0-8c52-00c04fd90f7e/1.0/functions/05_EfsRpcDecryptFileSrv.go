package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcDecryptFileSrvRequest carries the [in] parameters of EfsRpcDecryptFileSrv.
type efsRpcDecryptFileSrvRequest struct {
	FileName *ndr.WSTR `ndr:"unique"`
	OpenFlag ndr.DWORD
}

func (*efsRpcDecryptFileSrvRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcDecryptFileSrv }

// efsRpcDecryptFileSrvResponse carries the [out] parameters and return value of EfsRpcDecryptFileSrv.
type efsRpcDecryptFileSrvResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcDecryptFileSrv calls EfsRpcDecryptFileSrv (opnum 5) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcDecryptFileSrv(rpc ndr.Invoker, fileName *ndr.WSTR, openFlag ndr.DWORD) (err error) {
	req := &efsRpcDecryptFileSrvRequest{
		FileName: fileName,
		OpenFlag: openFlag,
	}
	var resp efsRpcDecryptFileSrvResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcDecryptFileSrv: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcDecryptFileSrv failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
