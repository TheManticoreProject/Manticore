package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcReadPrinterRequest carries the [in] parameters of RpcReadPrinter.
type rpcReadPrinterRequest struct {
	HPrinter structures.PRINTER_HANDLE
	CbBuf    ndr.DWORD
}

func (*rpcReadPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcReadPrinter }

// rpcReadPrinterResponse carries the [out] parameters and return value of RpcReadPrinter.
type rpcReadPrinterResponse struct {
	PBuf          []uint8 `ndr:"ref,size_is=CbBuf"`
	PcNoBytesRead ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcReadPrinter calls RpcReadPrinter (opnum 22) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcReadPrinter(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, cbBuf ndr.DWORD) (PBuf []uint8, PcNoBytesRead ndr.DWORD, err error) {
	req := &rpcReadPrinterRequest{
		HPrinter: hPrinter,
		CbBuf:    cbBuf,
	}
	var resp rpcReadPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcReadPrinter: %w", err)
		return
	}
	PBuf = resp.PBuf
	PcNoBytesRead = resp.PcNoBytesRead
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcReadPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
