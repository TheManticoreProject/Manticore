package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcEnumPrintProcessorDatatypesRequest carries the [in] parameters of RpcEnumPrintProcessorDatatypes.
type rpcEnumPrintProcessorDatatypesRequest struct {
	PName               *ndr.WSTR `ndr:"unique"`
	PPrintProcessorName *ndr.WSTR `ndr:"unique"`
	Level               ndr.DWORD
	PDatatypes          []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf               ndr.DWORD
}

func (*rpcEnumPrintProcessorDatatypesRequest) Opnum() uint16 {
	return winspool.OpnumRpcEnumPrintProcessorDatatypes
}

// rpcEnumPrintProcessorDatatypesResponse carries the [out] parameters and return value of RpcEnumPrintProcessorDatatypes.
type rpcEnumPrintProcessorDatatypesResponse struct {
	PDatatypes []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcEnumPrintProcessorDatatypes calls RpcEnumPrintProcessorDatatypes (opnum 51) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumPrintProcessorDatatypes(rpc ndr.Invoker, pName *ndr.WSTR, pPrintProcessorName *ndr.WSTR, level ndr.DWORD, pDatatypes []uint8, cbBuf ndr.DWORD) (PDatatypes []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcEnumPrintProcessorDatatypesRequest{
		PName:               pName,
		PPrintProcessorName: pPrintProcessorName,
		Level:               level,
		PDatatypes:          pDatatypes,
		CbBuf:               cbBuf,
	}
	var resp rpcEnumPrintProcessorDatatypesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumPrintProcessorDatatypes: %w", err)
		return
	}
	PDatatypes = resp.PDatatypes
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumPrintProcessorDatatypes failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
