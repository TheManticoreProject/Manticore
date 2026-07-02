package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAddMonitorRequest carries the [in] parameters of RpcAddMonitor.
type rpcAddMonitorRequest struct {
	Name              *ndr.WSTR `ndr:"unique"`
	PMonitorContainer structures.MONITOR_CONTAINER
}

func (*rpcAddMonitorRequest) Opnum() uint16 { return winspool.OpnumRpcAddMonitor }

// rpcAddMonitorResponse carries the [out] parameters and return value of RpcAddMonitor.
type rpcAddMonitorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAddMonitor calls RpcAddMonitor (opnum 46) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddMonitor(rpc ndr.Invoker, name *ndr.WSTR, pMonitorContainer structures.MONITOR_CONTAINER) (err error) {
	req := &rpcAddMonitorRequest{
		Name:              name,
		PMonitorContainer: pMonitorContainer,
	}
	var resp rpcAddMonitorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAddMonitor: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcAddMonitor failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
