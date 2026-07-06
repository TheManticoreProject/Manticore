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

// rpcEnumPrintProcessorsRequest carries the [in] parameters of RpcEnumPrintProcessors.
type rpcEnumPrintProcessorsRequest struct {
	PName               *ndr.WSTR `ndr:"unique"`
	PEnvironment        *ndr.WSTR `ndr:"unique"`
	Level               ndr.DWORD
	PPrintProcessorInfo []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf               ndr.DWORD
}

func (*rpcEnumPrintProcessorsRequest) Opnum() uint16 { return winspool.OpnumRpcEnumPrintProcessors }

// rpcEnumPrintProcessorsResponse carries the [out] parameters and return value of RpcEnumPrintProcessors.
type rpcEnumPrintProcessorsResponse struct {
	PPrintProcessorInfo []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded           ndr.DWORD
	PcReturned          ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// RpcEnumPrintProcessors calls RpcEnumPrintProcessors (opnum 15) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumPrintProcessors(rpc ndr.Invoker, pName *ndr.WSTR, pEnvironment *ndr.WSTR, level ndr.DWORD, pPrintProcessorInfo []uint8, cbBuf ndr.DWORD) (PPrintProcessorInfo []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcEnumPrintProcessorsRequest{
		PName:               pName,
		PEnvironment:        pEnvironment,
		Level:               level,
		PPrintProcessorInfo: pPrintProcessorInfo,
		CbBuf:               cbBuf,
	}
	var resp rpcEnumPrintProcessorsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumPrintProcessors: %w", err)
		return
	}
	PPrintProcessorInfo = resp.PPrintProcessorInfo
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumPrintProcessors failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
