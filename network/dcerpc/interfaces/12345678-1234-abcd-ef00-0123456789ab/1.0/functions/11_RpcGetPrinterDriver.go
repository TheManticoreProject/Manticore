package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetPrinterDriverRequest carries the [in] parameters of RpcGetPrinterDriver.
type rpcGetPrinterDriverRequest struct {
	HPrinter     structures.PRINTER_HANDLE
	PEnvironment *ndr.WSTR `ndr:"unique"`
	Level        ndr.DWORD
	PDriver      []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf        ndr.DWORD
}

func (*rpcGetPrinterDriverRequest) Opnum() uint16 { return winspool.OpnumRpcGetPrinterDriver }

// rpcGetPrinterDriverResponse carries the [out] parameters and return value of RpcGetPrinterDriver.
type rpcGetPrinterDriverResponse struct {
	PDriver   []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcGetPrinterDriver calls RpcGetPrinterDriver (opnum 11) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetPrinterDriver(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pEnvironment *ndr.WSTR, level ndr.DWORD, pDriver []uint8, cbBuf ndr.DWORD) (PDriver []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcGetPrinterDriverRequest{
		HPrinter:     hPrinter,
		PEnvironment: pEnvironment,
		Level:        level,
		PDriver:      pDriver,
		CbBuf:        cbBuf,
	}
	var resp rpcGetPrinterDriverResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetPrinterDriver: %w", err)
		return
	}
	PDriver = resp.PDriver
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetPrinterDriver failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
