package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncEnumPrintersRequest carries the [in] parameters of RpcAsyncEnumPrinters.
type rpcAsyncEnumPrintersRequest struct {
	Flags        ndr.DWORD
	Name         *ndr.WSTR `ndr:"unique"`
	Level        ndr.DWORD
	PPrinterEnum []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf        ndr.DWORD
}

func (*rpcAsyncEnumPrintersRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncEnumPrinters }

// rpcAsyncEnumPrintersResponse carries the [out] parameters and return value of RpcAsyncEnumPrinters.
type rpcAsyncEnumPrintersResponse struct {
	PPrinterEnum []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded    ndr.DWORD
	PcReturned   ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumPrinters calls RpcAsyncEnumPrinters (opnum 38) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumPrinters(rpc ndr.Invoker, flags ndr.DWORD, name *ndr.WSTR, level ndr.DWORD, pPrinterEnum []uint8, cbBuf ndr.DWORD) (PPrinterEnum []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcAsyncEnumPrintersRequest{
		Flags:        flags,
		Name:         name,
		Level:        level,
		PPrinterEnum: pPrinterEnum,
		CbBuf:        cbBuf,
	}
	var resp rpcAsyncEnumPrintersResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumPrinters: %w", err)
		return
	}
	PPrinterEnum = resp.PPrinterEnum
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumPrinters failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
