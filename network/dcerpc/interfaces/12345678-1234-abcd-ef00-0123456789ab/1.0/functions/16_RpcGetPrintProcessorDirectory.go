package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetPrintProcessorDirectoryRequest carries the [in] parameters of RpcGetPrintProcessorDirectory.
type rpcGetPrintProcessorDirectoryRequest struct {
	PName                    *ndr.WSTR `ndr:"unique"`
	PEnvironment             *ndr.WSTR `ndr:"unique"`
	Level                    ndr.DWORD
	PPrintProcessorDirectory []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf                    ndr.DWORD
}

func (*rpcGetPrintProcessorDirectoryRequest) Opnum() uint16 {
	return winspool.OpnumRpcGetPrintProcessorDirectory
}

// rpcGetPrintProcessorDirectoryResponse carries the [out] parameters and return value of RpcGetPrintProcessorDirectory.
type rpcGetPrintProcessorDirectoryResponse struct {
	PPrintProcessorDirectory []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded                ndr.DWORD
	Status                   ndr.DWORD `ndr:"retval"`
}

// RpcGetPrintProcessorDirectory calls RpcGetPrintProcessorDirectory (opnum 16) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetPrintProcessorDirectory(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment *ndr.WSTR, level ndr.DWORD, pPrintProcessorDirectory []uint8, cbBuf ndr.DWORD) (PPrintProcessorDirectory []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcGetPrintProcessorDirectoryRequest{
		PName:                    pName,
		PEnvironment:             pEnvironment,
		Level:                    level,
		PPrintProcessorDirectory: pPrintProcessorDirectory,
		CbBuf:                    cbBuf,
	}
	var resp rpcGetPrintProcessorDirectoryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetPrintProcessorDirectory: %w", err)
		return
	}
	PPrintProcessorDirectory = resp.PPrintProcessorDirectory
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetPrintProcessorDirectory failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
