package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcWritePrinterRequest carries the [in] parameters of RpcWritePrinter.
type rpcWritePrinterRequest struct {
	HPrinter structures.PRINTER_HANDLE
	PBuf     []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcWritePrinterRequest) Opnum() uint16 { return winspool.OpnumRpcWritePrinter }

// rpcWritePrinterResponse carries the [out] parameters and return value of RpcWritePrinter.
type rpcWritePrinterResponse struct {
	PcWritten ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcWritePrinter calls RpcWritePrinter (opnum 19) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcWritePrinter(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pBuf []uint8, cbBuf ndr.DWORD) (PcWritten ndr.DWORD, err error) {
	req := &rpcWritePrinterRequest{
		HPrinter: hPrinter,
		PBuf:     pBuf,
		CbBuf:    cbBuf,
	}
	var resp rpcWritePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWritePrinter: %w", err)
		return
	}
	PcWritten = resp.PcWritten
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcWritePrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
