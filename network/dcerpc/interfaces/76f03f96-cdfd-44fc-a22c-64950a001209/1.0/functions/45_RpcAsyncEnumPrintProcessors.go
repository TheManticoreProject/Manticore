package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncEnumPrintProcessorsRequest carries the [in] parameters of RpcAsyncEnumPrintProcessors.
type rpcAsyncEnumPrintProcessorsRequest struct {
	PName               *ndr.WSTR `ndr:"unique"`
	PEnvironment        *ndr.WSTR `ndr:"unique"`
	Level               ndr.DWORD
	PPrintProcessorInfo []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf               ndr.DWORD
}

func (*rpcAsyncEnumPrintProcessorsRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncEnumPrintProcessors
}

// rpcAsyncEnumPrintProcessorsResponse carries the [out] parameters and return value of RpcAsyncEnumPrintProcessors.
type rpcAsyncEnumPrintProcessorsResponse struct {
	PPrintProcessorInfo []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded           ndr.DWORD
	PcReturned          ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumPrintProcessors calls RpcAsyncEnumPrintProcessors (opnum 45) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumPrintProcessors(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment *ndr.WSTR, level ndr.DWORD, pPrintProcessorInfo []uint8, cbBuf ndr.DWORD) (PPrintProcessorInfo []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcAsyncEnumPrintProcessorsRequest{
		PName:               pName,
		PEnvironment:        pEnvironment,
		Level:               level,
		PPrintProcessorInfo: pPrintProcessorInfo,
		CbBuf:               cbBuf,
	}
	var resp rpcAsyncEnumPrintProcessorsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumPrintProcessors: %w", err)
		return
	}
	PPrintProcessorInfo = resp.PPrintProcessorInfo
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumPrintProcessors failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
