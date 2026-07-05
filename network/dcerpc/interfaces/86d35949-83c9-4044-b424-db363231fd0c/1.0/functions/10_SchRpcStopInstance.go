package functions

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// schRpcStopInstanceRequest carries the [in] parameters of SchRpcStopInstance.
type schRpcStopInstanceRequest struct {
	Guid  guid.GUID
	Flags ndr.DWORD
}

func (*schRpcStopInstanceRequest) Opnum() uint16 {
	return schrpc.OpnumSchRpcStopInstance
}

// schRpcStopInstanceResponse carries the [out] parameters and return value of SchRpcStopInstance.
type schRpcStopInstanceResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SchRpcStopInstance calls SchRpcStopInstance (opnum 10) ([MS-TSCH] section 3.2.5.4.11).
func SchRpcStopInstance(rpc ndr.Invoker, guid guid.GUID, flags ndr.DWORD) (err error) {
	req := &schRpcStopInstanceRequest{
		Guid:  guid,
		Flags: flags,
	}
	var resp schRpcStopInstanceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcStopInstance: %w", err)
		return
	}
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcStopInstance failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
