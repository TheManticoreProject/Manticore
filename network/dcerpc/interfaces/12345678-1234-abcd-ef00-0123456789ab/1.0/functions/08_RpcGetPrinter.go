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

// rpcGetPrinterRequest carries the [in] parameters of RpcGetPrinter.
type rpcGetPrinterRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
	Level    ndr.DWORD
	PPrinter []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcGetPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcGetPrinter }

// rpcGetPrinterResponse carries the [out] parameters and return value of RpcGetPrinter.
type rpcGetPrinterResponse struct {
	PPrinter  []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcGetPrinter calls RpcGetPrinter (opnum 8) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetPrinter(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, level ndr.DWORD, pPrinter []uint8, cbBuf ndr.DWORD) (PPrinter []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcGetPrinterRequest{
		HPrinter: hPrinter,
		Level:    level,
		PPrinter: pPrinter,
		CbBuf:    cbBuf,
	}
	var resp rpcGetPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetPrinter: %w", err)
		return
	}
	PPrinter = resp.PPrinter
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
