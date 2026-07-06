package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncGetPrintProcessorDirectoryRequest carries the [in] parameters of RpcAsyncGetPrintProcessorDirectory.
type rpcAsyncGetPrintProcessorDirectoryRequest struct {
	PName                    *ndr.WSTR `ndr:"unique"`
	PEnvironment             *ndr.WSTR `ndr:"unique"`
	Level                    ndr.DWORD
	PPrintProcessorDirectory []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf                    ndr.DWORD
}

func (*rpcAsyncGetPrintProcessorDirectoryRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncGetPrintProcessorDirectory
}

// rpcAsyncGetPrintProcessorDirectoryResponse carries the [out] parameters and return value of RpcAsyncGetPrintProcessorDirectory.
type rpcAsyncGetPrintProcessorDirectoryResponse struct {
	PPrintProcessorDirectory []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded                ndr.DWORD
	Status                   ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetPrintProcessorDirectory calls RpcAsyncGetPrintProcessorDirectory (opnum 46) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetPrintProcessorDirectory(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment *ndr.WSTR, level ndr.DWORD, pPrintProcessorDirectory []uint8, cbBuf ndr.DWORD) (PPrintProcessorDirectory []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcAsyncGetPrintProcessorDirectoryRequest{
		PName:                    pName,
		PEnvironment:             pEnvironment,
		Level:                    level,
		PPrintProcessorDirectory: pPrintProcessorDirectory,
		CbBuf:                    cbBuf,
	}
	var resp rpcAsyncGetPrintProcessorDirectoryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetPrintProcessorDirectory: %w", err)
		return
	}
	PPrintProcessorDirectory = resp.PPrintProcessorDirectory
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetPrintProcessorDirectory failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
