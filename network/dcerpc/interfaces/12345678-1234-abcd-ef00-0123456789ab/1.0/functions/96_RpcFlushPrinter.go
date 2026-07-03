package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcFlushPrinterRequest carries the [in] parameters of RpcFlushPrinter.
type rpcFlushPrinterRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
	PBuf     []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf    ndr.DWORD
	CSleep   ndr.DWORD
}

func (*rpcFlushPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcFlushPrinter }

// rpcFlushPrinterResponse carries the [out] parameters and return value of RpcFlushPrinter.
type rpcFlushPrinterResponse struct {
	PcWritten ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcFlushPrinter calls RpcFlushPrinter (opnum 96) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcFlushPrinter(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pBuf []uint8, cbBuf ndr.DWORD, cSleep ndr.DWORD) (PcWritten ndr.DWORD, err error) {
	req := &rpcFlushPrinterRequest{
		HPrinter: hPrinter,
		PBuf:     pBuf,
		CbBuf:    cbBuf,
		CSleep:   cSleep,
	}
	var resp rpcFlushPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcFlushPrinter: %w", err)
		return
	}
	PcWritten = resp.PcWritten
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcFlushPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
