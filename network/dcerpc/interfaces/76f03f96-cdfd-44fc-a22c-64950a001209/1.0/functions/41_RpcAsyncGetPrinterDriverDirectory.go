package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncGetPrinterDriverDirectoryRequest carries the [in] parameters of RpcAsyncGetPrinterDriverDirectory.
type rpcAsyncGetPrinterDriverDirectoryRequest struct {
	PName            *ndr.WSTR `ndr:"unique"`
	PEnvironment     *ndr.WSTR `ndr:"unique"`
	Level            ndr.DWORD
	PDriverDirectory []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf            ndr.DWORD
}

func (*rpcAsyncGetPrinterDriverDirectoryRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncGetPrinterDriverDirectory
}

// rpcAsyncGetPrinterDriverDirectoryResponse carries the [out] parameters and return value of RpcAsyncGetPrinterDriverDirectory.
type rpcAsyncGetPrinterDriverDirectoryResponse struct {
	PDriverDirectory []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded        ndr.DWORD
	Status           ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetPrinterDriverDirectory calls RpcAsyncGetPrinterDriverDirectory (opnum 41) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetPrinterDriverDirectory(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment *ndr.WSTR, level ndr.DWORD, pDriverDirectory []uint8, cbBuf ndr.DWORD) (PDriverDirectory []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcAsyncGetPrinterDriverDirectoryRequest{
		PName:            pName,
		PEnvironment:     pEnvironment,
		Level:            level,
		PDriverDirectory: pDriverDirectory,
		CbBuf:            cbBuf,
	}
	var resp rpcAsyncGetPrinterDriverDirectoryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetPrinterDriverDirectory: %w", err)
		return
	}
	PDriverDirectory = resp.PDriverDirectory
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetPrinterDriverDirectory failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
