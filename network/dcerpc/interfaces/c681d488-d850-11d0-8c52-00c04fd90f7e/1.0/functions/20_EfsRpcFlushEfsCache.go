package functions

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// efsRpcFlushEfsCacheRequest carries the [in] parameters of EfsRpcFlushEfsCache.
type efsRpcFlushEfsCacheRequest struct {
}

func (*efsRpcFlushEfsCacheRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcFlushEfsCache }

// efsRpcFlushEfsCacheResponse carries the [out] parameters and return value of EfsRpcFlushEfsCache.
type efsRpcFlushEfsCacheResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EfsRpcFlushEfsCache calls EfsRpcFlushEfsCache (opnum 20) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcFlushEfsCache(rpc ndr.Invoker) (err error) {
	req := &efsRpcFlushEfsCacheRequest{}
	var resp efsRpcFlushEfsCacheResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcFlushEfsCache: %w", err)
		return
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcFlushEfsCache failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
