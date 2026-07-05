package functions

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// schRpcDeleteRequest carries the [in] parameters of SchRpcDelete.
type schRpcDeleteRequest struct {
	Path  ndr.WSTR
	Flags ndr.DWORD
}

func (*schRpcDeleteRequest) Opnum() uint16 { return schrpc.OpnumSchRpcDelete }

// schRpcDeleteResponse carries the [out] parameters and return value of SchRpcDelete.
type schRpcDeleteResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SchRpcDelete calls SchRpcDelete (opnum 13) ([MS-TSCH] section 3.2.5.4.14).
func SchRpcDelete(rpc ndr.Invoker, path ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &schRpcDeleteRequest{
		Path:  path,
		Flags: flags,
	}
	var resp schRpcDeleteResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcDelete: %w", err)
		return
	}
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcDelete failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
