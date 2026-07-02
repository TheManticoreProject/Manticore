package functions

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSLookupEndRequest carries the [in] parameters of S_DSLookupEnd.
type s_DSLookupEndRequest struct {
	PhContext msmqds.PPCONTEXT_HANDLE_TYPE
}

func (*s_DSLookupEndRequest) Opnum() uint16 { return dscomm.OpnumS_DSLookupEnd }

// s_DSLookupEndResponse carries the [out] parameters and return value of S_DSLookupEnd.
type s_DSLookupEndResponse struct {
	PhContext msmqds.PPCONTEXT_HANDLE_TYPE
	Status    ndr.DWORD `ndr:"retval"`
}

// S_DSLookupEnd calls S_DSLookupEnd (opnum 8) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSLookupEnd(rpc ndr.Invoker, phContext msmqds.PPCONTEXT_HANDLE_TYPE) (PhContext msmqds.PPCONTEXT_HANDLE_TYPE, err error) {
	req := &s_DSLookupEndRequest{
		PhContext: phContext,
	}
	var resp s_DSLookupEndResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSLookupEnd: %w", err)
		return
	}
	PhContext = resp.PhContext
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSLookupEnd failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}
