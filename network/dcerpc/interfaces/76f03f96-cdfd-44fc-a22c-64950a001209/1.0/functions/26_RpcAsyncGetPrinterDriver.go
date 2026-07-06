package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncGetPrinterDriverRequest carries the [in] parameters of RpcAsyncGetPrinterDriver.
type rpcAsyncGetPrinterDriverRequest struct {
	HPrinter             mspar.PRINTER_HANDLE
	PEnvironment         *ndr.WSTR `ndr:"unique"`
	Level                ndr.DWORD
	PDriver              []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf                ndr.DWORD
	DwClientMajorVersion ndr.DWORD
	DwClientMinorVersion ndr.DWORD
}

func (*rpcAsyncGetPrinterDriverRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncGetPrinterDriver
}

// rpcAsyncGetPrinterDriverResponse carries the [out] parameters and return value of RpcAsyncGetPrinterDriver.
type rpcAsyncGetPrinterDriverResponse struct {
	PDriver             []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded           ndr.DWORD
	PdwServerMaxVersion ndr.DWORD
	PdwServerMinVersion ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetPrinterDriver calls RpcAsyncGetPrinterDriver (opnum 26) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetPrinterDriver(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pEnvironment *ndr.WSTR, level ndr.DWORD, pDriver []uint8, cbBuf ndr.DWORD, dwClientMajorVersion ndr.DWORD, dwClientMinorVersion ndr.DWORD) (PDriver []uint8, PcbNeeded ndr.DWORD, PdwServerMaxVersion ndr.DWORD, PdwServerMinVersion ndr.DWORD, err error) {
	req := &rpcAsyncGetPrinterDriverRequest{
		HPrinter:             hPrinter,
		PEnvironment:         pEnvironment,
		Level:                level,
		PDriver:              pDriver,
		CbBuf:                cbBuf,
		DwClientMajorVersion: dwClientMajorVersion,
		DwClientMinorVersion: dwClientMinorVersion,
	}
	var resp rpcAsyncGetPrinterDriverResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetPrinterDriver: %w", err)
		return
	}
	PDriver = resp.PDriver
	PcbNeeded = resp.PcbNeeded
	PdwServerMaxVersion = resp.PdwServerMaxVersion
	PdwServerMinVersion = resp.PdwServerMinVersion
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetPrinterDriver failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
