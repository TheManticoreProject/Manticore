package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcDeleteMonitorRequest carries the [in] parameters of RpcDeleteMonitor.
type rpcDeleteMonitorRequest struct {
	Name         *ndr.WSTR `ndr:"unique"`
	PEnvironment *ndr.WSTR `ndr:"unique"`
	PMonitorName ndr.WSTR
}

func (*rpcDeleteMonitorRequest) Opnum() uint16 { return winspool.OpnumRpcDeleteMonitor }

// rpcDeleteMonitorResponse carries the [out] parameters and return value of RpcDeleteMonitor.
type rpcDeleteMonitorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeleteMonitor calls RpcDeleteMonitor (opnum 47) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeleteMonitor(rpc ndr.Invoker, name *ndr.WSTR, pEnvironment *ndr.WSTR, pMonitorName ndr.WSTR) (err error) {
	req := &rpcDeleteMonitorRequest{
		Name:         name,
		PEnvironment: pEnvironment,
		PMonitorName: pMonitorName,
	}
	var resp rpcDeleteMonitorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeleteMonitor: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeleteMonitor failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
