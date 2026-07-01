package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetPrinterDriverDirectoryRequest carries the [in] parameters of RpcGetPrinterDriverDirectory.
type rpcGetPrinterDriverDirectoryRequest struct {
	PName            *ndr.WSTR `ndr:"unique"`
	PEnvironment     *ndr.WSTR `ndr:"unique"`
	Level            ndr.DWORD
	PDriverDirectory []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf            ndr.DWORD
}

func (*rpcGetPrinterDriverDirectoryRequest) Opnum() uint16 {
	return winspool.OpnumRpcGetPrinterDriverDirectory
}

// rpcGetPrinterDriverDirectoryResponse carries the [out] parameters and return value of RpcGetPrinterDriverDirectory.
type rpcGetPrinterDriverDirectoryResponse struct {
	PDriverDirectory []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded        ndr.DWORD
	Status           ndr.DWORD `ndr:"retval"`
}

// RpcGetPrinterDriverDirectory calls RpcGetPrinterDriverDirectory (opnum 12) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetPrinterDriverDirectory(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment *ndr.WSTR, level ndr.DWORD, pDriverDirectory []uint8, cbBuf ndr.DWORD) (PDriverDirectory []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcGetPrinterDriverDirectoryRequest{
		PName:            pName,
		PEnvironment:     pEnvironment,
		Level:            level,
		PDriverDirectory: pDriverDirectory,
		CbBuf:            cbBuf,
	}
	var resp rpcGetPrinterDriverDirectoryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetPrinterDriverDirectory: %w", err)
		return
	}
	PDriverDirectory = resp.PDriverDirectory
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetPrinterDriverDirectory failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
