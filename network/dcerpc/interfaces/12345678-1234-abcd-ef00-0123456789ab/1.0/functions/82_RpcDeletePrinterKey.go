package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcDeletePrinterKeyRequest carries the [in] parameters of RpcDeletePrinterKey.
type rpcDeletePrinterKeyRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
	PKeyName ndr.WSTR
}

func (*rpcDeletePrinterKeyRequest) Opnum() uint16 { return winspool.OpnumRpcDeletePrinterKey }

// rpcDeletePrinterKeyResponse carries the [out] parameters and return value of RpcDeletePrinterKey.
type rpcDeletePrinterKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeletePrinterKey calls RpcDeletePrinterKey (opnum 82) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePrinterKey(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pKeyName ndr.WSTR) (err error) {
	req := &rpcDeletePrinterKeyRequest{
		HPrinter: hPrinter,
		PKeyName: pKeyName,
	}
	var resp rpcDeletePrinterKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeletePrinterKey: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeletePrinterKey failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
