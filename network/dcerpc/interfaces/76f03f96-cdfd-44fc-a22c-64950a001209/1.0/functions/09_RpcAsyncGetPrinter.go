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

// rpcAsyncGetPrinterRequest carries the [in] parameters of RpcAsyncGetPrinter.
type rpcAsyncGetPrinterRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	Level    ndr.DWORD
	PPrinter []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcAsyncGetPrinterRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncGetPrinter }

// rpcAsyncGetPrinterResponse carries the [out] parameters and return value of RpcAsyncGetPrinter.
type rpcAsyncGetPrinterResponse struct {
	PPrinter  []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetPrinter calls RpcAsyncGetPrinter (opnum 9) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetPrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, level ndr.DWORD, pPrinter []uint8, cbBuf ndr.DWORD) (PPrinter []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcAsyncGetPrinterRequest{
		HPrinter: hPrinter,
		Level:    level,
		PPrinter: pPrinter,
		CbBuf:    cbBuf,
	}
	var resp rpcAsyncGetPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetPrinter: %w", err)
		return
	}
	PPrinter = resp.PPrinter
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetPrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
