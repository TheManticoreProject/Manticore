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

// rpcRemoteFindFirstPrinterChangeNotificationRequest carries the [in] parameters of RpcRemoteFindFirstPrinterChangeNotification.
type rpcRemoteFindFirstPrinterChangeNotificationRequest struct {
	HPrinter        msrprn.PRINTER_HANDLE
	FdwFlags        ndr.DWORD
	FdwOptions      ndr.DWORD
	PszLocalMachine *ndr.WSTR `ndr:"unique"`
	DwPrinterLocal  ndr.DWORD
	CbBuffer        ndr.DWORD
	PBuffer         []uint8 `ndr:"unique,size_is=CbBuffer"`
}

func (*rpcRemoteFindFirstPrinterChangeNotificationRequest) Opnum() uint16 {
	return winspool.OpnumRpcRemoteFindFirstPrinterChangeNotification
}

// rpcRemoteFindFirstPrinterChangeNotificationResponse carries the [out] parameters and return value of RpcRemoteFindFirstPrinterChangeNotification.
type rpcRemoteFindFirstPrinterChangeNotificationResponse struct {
	PBuffer []uint8   `ndr:"unique,size_is=CbBuffer"`
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcRemoteFindFirstPrinterChangeNotification calls RpcRemoteFindFirstPrinterChangeNotification (opnum 62) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcRemoteFindFirstPrinterChangeNotification(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, fdwFlags ndr.DWORD, fdwOptions ndr.DWORD, pszLocalMachine *ndr.WSTR, dwPrinterLocal ndr.DWORD, cbBuffer ndr.DWORD, pBuffer []uint8) (PBuffer []uint8, err error) {
	req := &rpcRemoteFindFirstPrinterChangeNotificationRequest{
		HPrinter:        hPrinter,
		FdwFlags:        fdwFlags,
		FdwOptions:      fdwOptions,
		PszLocalMachine: pszLocalMachine,
		DwPrinterLocal:  dwPrinterLocal,
		CbBuffer:        cbBuffer,
		PBuffer:         pBuffer,
	}
	var resp rpcRemoteFindFirstPrinterChangeNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcRemoteFindFirstPrinterChangeNotification: %w", err)
		return
	}
	PBuffer = resp.PBuffer
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcRemoteFindFirstPrinterChangeNotification failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
