package functions

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// schRpcEnumInstancesRequest carries the [in] parameters of SchRpcEnumInstances.
type schRpcEnumInstancesRequest struct {
	Path  *ndr.WSTR `ndr:"unique"`
	Flags ndr.DWORD
}

func (*schRpcEnumInstancesRequest) Opnum() uint16 {
	return schrpc.OpnumSchRpcEnumInstances
}

// schRpcEnumInstancesResponse carries the [out] parameters and return value of SchRpcEnumInstances.
type schRpcEnumInstancesResponse struct {
	PcGuids ndr.DWORD
	PGuids  []guid.GUID `ndr:"unique,size_is=PcGuids"`
	Status  ndr.DWORD   `ndr:"retval"`
}

// SchRpcEnumInstances calls SchRpcEnumInstances (opnum 8) ([MS-TSCH] section 3.2.5.4.9).
func SchRpcEnumInstances(rpc ndr.Invoker, path *ndr.WSTR, flags ndr.DWORD) (PcGuids ndr.DWORD, PGuids []guid.GUID, err error) {
	req := &schRpcEnumInstancesRequest{
		Path:  path,
		Flags: flags,
	}
	var resp schRpcEnumInstancesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcEnumInstances: %w", err)
		return
	}
	PcGuids = resp.PcGuids
	PGuids = resp.PGuids
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcEnumInstances failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
