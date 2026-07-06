package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

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
