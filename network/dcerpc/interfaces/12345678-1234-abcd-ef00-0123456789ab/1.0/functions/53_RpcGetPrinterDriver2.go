package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcGetPrinterDriver2Request carries the [in] parameters of RpcGetPrinterDriver2.
type rpcGetPrinterDriver2Request struct {
	HPrinter             msrprn.PRINTER_HANDLE
	PEnvironment         *ndr.WSTR `ndr:"unique"`
	Level                ndr.DWORD
	PDriver              []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf                ndr.DWORD
	DwClientMajorVersion ndr.DWORD
	DwClientMinorVersion ndr.DWORD
}

func (*rpcGetPrinterDriver2Request) Opnum() uint16 { return winspool.OpnumRpcGetPrinterDriver2 }

// rpcGetPrinterDriver2Response carries the [out] parameters and return value of RpcGetPrinterDriver2.
type rpcGetPrinterDriver2Response struct {
	PDriver             []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded           ndr.DWORD
	PdwServerMaxVersion ndr.DWORD
	PdwServerMinVersion ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// RpcGetPrinterDriver2 calls RpcGetPrinterDriver2 (opnum 53) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetPrinterDriver2(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pEnvironment *ndr.WSTR, level ndr.DWORD, pDriver []uint8, cbBuf ndr.DWORD, dwClientMajorVersion ndr.DWORD, dwClientMinorVersion ndr.DWORD) (PDriver []uint8, PcbNeeded ndr.DWORD, PdwServerMaxVersion ndr.DWORD, PdwServerMinVersion ndr.DWORD, err error) {
	req := &rpcGetPrinterDriver2Request{
		HPrinter:             hPrinter,
		PEnvironment:         pEnvironment,
		Level:                level,
		PDriver:              pDriver,
		CbBuf:                cbBuf,
		DwClientMajorVersion: dwClientMajorVersion,
		DwClientMinorVersion: dwClientMinorVersion,
	}
	var resp rpcGetPrinterDriver2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetPrinterDriver2: %w", err)
		return
	}
	PDriver = resp.PDriver
	PcbNeeded = resp.PcbNeeded
	PdwServerMaxVersion = resp.PdwServerMaxVersion
	PdwServerMinVersion = resp.PdwServerMinVersion
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetPrinterDriver2 failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
