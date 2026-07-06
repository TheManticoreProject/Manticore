package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetSessionCountersRequest carries the [in] parameters of RpcGetSessionCounters.
type rpcGetSessionCountersRequest struct {
	PCounter []mststs.TS_COUNTER `ndr:"ref,size_is=UEntries"`
	UEntries ndr.DWORD
}

func (*rpcGetSessionCountersRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcGetSessionCounters }

// rpcGetSessionCountersResponse carries the [out] parameters and return value of RpcGetSessionCounters.
type rpcGetSessionCountersResponse struct {
	PCounter []mststs.TS_COUNTER `ndr:"ref,size_is=UEntries"`
	Status   ndr.DWORD           `ndr:"retval"`
}

// RpcGetSessionCounters calls RpcGetSessionCounters (opnum 11) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetSessionCounters(rpc ndr.Invoker, pCounter []mststs.TS_COUNTER, uEntries ndr.DWORD) (PCounter []mststs.TS_COUNTER, err error) {
	req := &rpcGetSessionCountersRequest{
		PCounter: pCounter,
		UEntries: uEntries,
	}
	var resp rpcGetSessionCountersResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetSessionCounters: %w", err)
		return
	}
	PCounter = resp.PCounter
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcGetSessionCounters failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
