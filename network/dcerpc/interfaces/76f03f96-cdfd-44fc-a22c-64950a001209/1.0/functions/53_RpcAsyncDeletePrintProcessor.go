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

// rpcAsyncDeletePrintProcessorRequest carries the [in] parameters of RpcAsyncDeletePrintProcessor.
type rpcAsyncDeletePrintProcessorRequest struct {
	Name                *ndr.WSTR `ndr:"unique"`
	PEnvironment        *ndr.WSTR `ndr:"unique"`
	PPrintProcessorName ndr.WSTR
}

func (*rpcAsyncDeletePrintProcessorRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeletePrintProcessor
}

// rpcAsyncDeletePrintProcessorResponse carries the [out] parameters and return value of RpcAsyncDeletePrintProcessor.
type rpcAsyncDeletePrintProcessorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeletePrintProcessor calls RpcAsyncDeletePrintProcessor (opnum 53) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeletePrintProcessor(rpc ndr.Invoker, name *ndr.WSTR, pEnvironment *ndr.WSTR, pPrintProcessorName ndr.WSTR) (err error) {
	req := &rpcAsyncDeletePrintProcessorRequest{
		Name:                name,
		PEnvironment:        pEnvironment,
		PPrintProcessorName: pPrintProcessorName,
	}
	var resp rpcAsyncDeletePrintProcessorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeletePrintProcessor: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeletePrintProcessor failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
