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

// rpcAsyncDeletePrinterKeyRequest carries the [in] parameters of RpcAsyncDeletePrinterKey.
type rpcAsyncDeletePrinterKeyRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	PKeyName ndr.WSTR
}

func (*rpcAsyncDeletePrinterKeyRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeletePrinterKey
}

// rpcAsyncDeletePrinterKeyResponse carries the [out] parameters and return value of RpcAsyncDeletePrinterKey.
type rpcAsyncDeletePrinterKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeletePrinterKey calls RpcAsyncDeletePrinterKey (opnum 32) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeletePrinterKey(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pKeyName ndr.WSTR) (err error) {
	req := &rpcAsyncDeletePrinterKeyRequest{
		HPrinter: hPrinter,
		PKeyName: pKeyName,
	}
	var resp rpcAsyncDeletePrinterKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeletePrinterKey: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeletePrinterKey failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
