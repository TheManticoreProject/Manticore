package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncEnumPrintProcessorDatatypesRequest carries the [in] parameters of RpcAsyncEnumPrintProcessorDatatypes.
type rpcAsyncEnumPrintProcessorDatatypesRequest struct {
	PName               *ndr.WSTR `ndr:"unique"`
	PPrintProcessorName *ndr.WSTR `ndr:"unique"`
	Level               ndr.DWORD
	PDatatypes          []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf               ndr.DWORD
}

func (*rpcAsyncEnumPrintProcessorDatatypesRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncEnumPrintProcessorDatatypes
}

// rpcAsyncEnumPrintProcessorDatatypesResponse carries the [out] parameters and return value of RpcAsyncEnumPrintProcessorDatatypes.
type rpcAsyncEnumPrintProcessorDatatypesResponse struct {
	PDatatypes []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumPrintProcessorDatatypes calls RpcAsyncEnumPrintProcessorDatatypes (opnum 54) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumPrintProcessorDatatypes(rpc ndr.Invoker, pName *ndr.WSTR, pPrintProcessorName *ndr.WSTR, level ndr.DWORD, pDatatypes []uint8, cbBuf ndr.DWORD) (PDatatypes []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcAsyncEnumPrintProcessorDatatypesRequest{
		PName:               pName,
		PPrintProcessorName: pPrintProcessorName,
		Level:               level,
		PDatatypes:          pDatatypes,
		CbBuf:               cbBuf,
	}
	var resp rpcAsyncEnumPrintProcessorDatatypesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumPrintProcessorDatatypes: %w", err)
		return
	}
	PDatatypes = resp.PDatatypes
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumPrintProcessorDatatypes failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
