package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAddPrintProcessorRequest carries the [in] parameters of RpcAddPrintProcessor.
type rpcAddPrintProcessorRequest struct {
	PName               *ndr.WSTR `ndr:"unique"`
	PEnvironment        ndr.WSTR
	PPathName           ndr.WSTR
	PPrintProcessorName ndr.WSTR
}

func (*rpcAddPrintProcessorRequest) Opnum() uint16 { return winspool.OpnumRpcAddPrintProcessor }

// rpcAddPrintProcessorResponse carries the [out] parameters and return value of RpcAddPrintProcessor.
type rpcAddPrintProcessorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAddPrintProcessor calls RpcAddPrintProcessor (opnum 14) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddPrintProcessor(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment ndr.WSTR, pPathName ndr.WSTR, pPrintProcessorName ndr.WSTR) (err error) {
	req := &rpcAddPrintProcessorRequest{
		PName:               pName,
		PEnvironment:        pEnvironment,
		PPathName:           pPathName,
		PPrintProcessorName: pPrintProcessorName,
	}
	var resp rpcAddPrintProcessorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAddPrintProcessor: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcAddPrintProcessor failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
