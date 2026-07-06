package functions

// IDL source: [MS-SWN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-swn/ccebaef8-60b0-4847-9ed7-2519d2b6ef19
// A fetched copy is kept at ms-swn.idl in the interface directory.

import (
	"fmt"

	Witness "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ccd8c074-d0e5-4a40-92b4-d074faa6ba28/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msswn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-swn"
)

// witnessrAsyncNotifyRequest carries the [in] parameters of WitnessrAsyncNotify.
type witnessrAsyncNotifyRequest struct {
	PContext msswn.PCONTEXT_HANDLE_SHARED
}

func (*witnessrAsyncNotifyRequest) Opnum() uint16 { return Witness.OpnumWitnessrAsyncNotify }

// witnessrAsyncNotifyResponse carries the [out] parameters and return value of WitnessrAsyncNotify.
type witnessrAsyncNotifyResponse struct {
	PResp  *msswn.RESP_ASYNC_NOTIFY `ndr:"unique"`
	Status ndr.DWORD                `ndr:"retval"`
}

// WitnessrAsyncNotify calls WitnessrAsyncNotify (opnum 3) ([MS-SWN] 3.1.4.4). It long-polls the server for pending notifications on a registration; the server blocks until a notification is available.
func WitnessrAsyncNotify(rpc ndr.Invoker, pContext msswn.PCONTEXT_HANDLE_SHARED) (PResp *msswn.RESP_ASYNC_NOTIFY, err error) {
	req := &witnessrAsyncNotifyRequest{
		PContext: pContext,
	}
	var resp witnessrAsyncNotifyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("WitnessrAsyncNotify: %w", err)
		return
	}
	PResp = resp.PResp
	if uint32(resp.Status) != Witness.StatusSuccess {
		err = fmt.Errorf("WitnessrAsyncNotify failed: %s", Witness.StatusString(uint32(resp.Status)))
	}
	return
}
