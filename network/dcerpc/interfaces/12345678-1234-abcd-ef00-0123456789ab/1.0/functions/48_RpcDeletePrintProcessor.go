package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcDeletePrintProcessorRequest carries the [in] parameters of RpcDeletePrintProcessor.
type rpcDeletePrintProcessorRequest struct {
	Name                *ndr.WSTR `ndr:"unique"`
	PEnvironment        *ndr.WSTR `ndr:"unique"`
	PPrintProcessorName ndr.WSTR
}

func (*rpcDeletePrintProcessorRequest) Opnum() uint16 { return winspool.OpnumRpcDeletePrintProcessor }

// rpcDeletePrintProcessorResponse carries the [out] parameters and return value of RpcDeletePrintProcessor.
type rpcDeletePrintProcessorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeletePrintProcessor calls RpcDeletePrintProcessor (opnum 48) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePrintProcessor(rpc ndr.Invoker, name *ndr.WSTR, pEnvironment *ndr.WSTR, pPrintProcessorName ndr.WSTR) (err error) {
	req := &rpcDeletePrintProcessorRequest{
		Name:                name,
		PEnvironment:        pEnvironment,
		PPrintProcessorName: pPrintProcessorName,
	}
	var resp rpcDeletePrintProcessorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeletePrintProcessor: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeletePrintProcessor failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
