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

// rpcRemoteFindFirstPrinterChangeNotificationExRequest carries the [in] parameters of RpcRemoteFindFirstPrinterChangeNotificationEx.
type rpcRemoteFindFirstPrinterChangeNotificationExRequest struct {
	HPrinter        msrprn.PRINTER_HANDLE
	FdwFlags        ndr.DWORD
	FdwOptions      ndr.DWORD
	PszLocalMachine *ndr.WSTR `ndr:"unique"`
	DwPrinterLocal  ndr.DWORD
	POptions        *msrprn.RPC_V2_NOTIFY_OPTIONS `ndr:"unique"`
}

func (*rpcRemoteFindFirstPrinterChangeNotificationExRequest) Opnum() uint16 {
	return winspool.OpnumRpcRemoteFindFirstPrinterChangeNotificationEx
}

// rpcRemoteFindFirstPrinterChangeNotificationExResponse carries the [out] parameters and return value of RpcRemoteFindFirstPrinterChangeNotificationEx.
type rpcRemoteFindFirstPrinterChangeNotificationExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcRemoteFindFirstPrinterChangeNotificationEx calls RpcRemoteFindFirstPrinterChangeNotificationEx (opnum 65) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcRemoteFindFirstPrinterChangeNotificationEx(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, fdwFlags ndr.DWORD, fdwOptions ndr.DWORD, pszLocalMachine *ndr.WSTR, dwPrinterLocal ndr.DWORD, pOptions *msrprn.RPC_V2_NOTIFY_OPTIONS) (err error) {
	req := &rpcRemoteFindFirstPrinterChangeNotificationExRequest{
		HPrinter:        hPrinter,
		FdwFlags:        fdwFlags,
		FdwOptions:      fdwOptions,
		PszLocalMachine: pszLocalMachine,
		DwPrinterLocal:  dwPrinterLocal,
		POptions:        pOptions,
	}
	var resp rpcRemoteFindFirstPrinterChangeNotificationExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcRemoteFindFirstPrinterChangeNotificationEx: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcRemoteFindFirstPrinterChangeNotificationEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
