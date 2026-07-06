package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncDeleteMonitorRequest carries the [in] parameters of RpcAsyncDeleteMonitor.
type rpcAsyncDeleteMonitorRequest struct {
	Name         *ndr.WSTR `ndr:"unique"`
	PEnvironment *ndr.WSTR `ndr:"unique"`
	PMonitorName ndr.WSTR
}

func (*rpcAsyncDeleteMonitorRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeleteMonitor
}

// rpcAsyncDeleteMonitorResponse carries the [out] parameters and return value of RpcAsyncDeleteMonitor.
type rpcAsyncDeleteMonitorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeleteMonitor calls RpcAsyncDeleteMonitor (opnum 52) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeleteMonitor(rpc ndr.Invoker, name *ndr.WSTR, pEnvironment *ndr.WSTR, pMonitorName ndr.WSTR) (err error) {
	req := &rpcAsyncDeleteMonitorRequest{
		Name:         name,
		PEnvironment: pEnvironment,
		PMonitorName: pMonitorName,
	}
	var resp rpcAsyncDeleteMonitorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeleteMonitor: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeleteMonitor failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
