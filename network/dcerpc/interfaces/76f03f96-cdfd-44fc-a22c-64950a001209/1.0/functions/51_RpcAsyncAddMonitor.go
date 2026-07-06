package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncAddMonitorRequest carries the [in] parameters of RpcAsyncAddMonitor.
type rpcAsyncAddMonitorRequest struct {
	Name              *ndr.WSTR `ndr:"unique"`
	PMonitorContainer mspar.MONITOR_CONTAINER
}

func (*rpcAsyncAddMonitorRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncAddMonitor }

// rpcAsyncAddMonitorResponse carries the [out] parameters and return value of RpcAsyncAddMonitor.
type rpcAsyncAddMonitorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncAddMonitor calls RpcAsyncAddMonitor (opnum 51) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncAddMonitor(rpc ndr.Invoker, name *ndr.WSTR, pMonitorContainer mspar.MONITOR_CONTAINER) (err error) {
	req := &rpcAsyncAddMonitorRequest{
		Name:              name,
		PMonitorContainer: pMonitorContainer,
	}
	var resp rpcAsyncAddMonitorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncAddMonitor: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncAddMonitor failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
