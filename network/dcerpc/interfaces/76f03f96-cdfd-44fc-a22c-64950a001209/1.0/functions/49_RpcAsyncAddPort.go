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

// rpcAsyncAddPortRequest carries the [in] parameters of RpcAsyncAddPort.
type rpcAsyncAddPortRequest struct {
	PName             *ndr.WSTR `ndr:"unique"`
	PPortContainer    mspar.PORT_CONTAINER
	PPortVarContainer mspar.PORT_VAR_CONTAINER
	PMonitorName      ndr.WSTR
}

func (*rpcAsyncAddPortRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncAddPort }

// rpcAsyncAddPortResponse carries the [out] parameters and return value of RpcAsyncAddPort.
type rpcAsyncAddPortResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncAddPort calls RpcAsyncAddPort (opnum 49) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncAddPort(rpc ndr.Invoker, pName *ndr.WSTR, pPortContainer mspar.PORT_CONTAINER, pPortVarContainer mspar.PORT_VAR_CONTAINER, pMonitorName ndr.WSTR) (err error) {
	req := &rpcAsyncAddPortRequest{
		PName:             pName,
		PPortContainer:    pPortContainer,
		PPortVarContainer: pPortVarContainer,
		PMonitorName:      pMonitorName,
	}
	var resp rpcAsyncAddPortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncAddPort: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncAddPort failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
