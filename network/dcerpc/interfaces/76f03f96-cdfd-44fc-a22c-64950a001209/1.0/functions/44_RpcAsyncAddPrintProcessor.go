package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncAddPrintProcessorRequest carries the [in] parameters of RpcAsyncAddPrintProcessor.
type rpcAsyncAddPrintProcessorRequest struct {
	PName               *ndr.WSTR `ndr:"unique"`
	PEnvironment        ndr.WSTR
	PPathName           ndr.WSTR
	PPrintProcessorName ndr.WSTR
}

func (*rpcAsyncAddPrintProcessorRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncAddPrintProcessor
}

// rpcAsyncAddPrintProcessorResponse carries the [out] parameters and return value of RpcAsyncAddPrintProcessor.
type rpcAsyncAddPrintProcessorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncAddPrintProcessor calls RpcAsyncAddPrintProcessor (opnum 44) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncAddPrintProcessor(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment ndr.WSTR, pPathName ndr.WSTR, pPrintProcessorName ndr.WSTR) (err error) {
	req := &rpcAsyncAddPrintProcessorRequest{
		PName:               pName,
		PEnvironment:        pEnvironment,
		PPathName:           pPathName,
		PPrintProcessorName: pPrintProcessorName,
	}
	var resp rpcAsyncAddPrintProcessorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncAddPrintProcessor: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncAddPrintProcessor failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
