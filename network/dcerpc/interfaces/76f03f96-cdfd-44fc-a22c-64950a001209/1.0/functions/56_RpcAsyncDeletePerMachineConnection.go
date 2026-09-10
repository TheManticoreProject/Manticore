package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// rpcAsyncDeletePerMachineConnectionRequest carries the [in] parameters of RpcAsyncDeletePerMachineConnection.
type rpcAsyncDeletePerMachineConnectionRequest struct {
	PServer      *ndr.WSTR `ndr:"unique"`
	PPrinterName ndr.WSTR
}

func (*rpcAsyncDeletePerMachineConnectionRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeletePerMachineConnection
}

// rpcAsyncDeletePerMachineConnectionResponse carries the [out] parameters and return value of RpcAsyncDeletePerMachineConnection.
type rpcAsyncDeletePerMachineConnectionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeletePerMachineConnection calls RpcAsyncDeletePerMachineConnection (opnum 56) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeletePerMachineConnection(rpc ndr.Invoker, pServer *ndr.WSTR, pPrinterName ndr.WSTR) (err error) {
	req := &rpcAsyncDeletePerMachineConnectionRequest{
		PServer:      pServer,
		PPrinterName: pPrinterName,
	}
	var resp rpcAsyncDeletePerMachineConnectionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeletePerMachineConnection: %w", err)
		return
	}
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("RpcAsyncDeletePerMachineConnection failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
