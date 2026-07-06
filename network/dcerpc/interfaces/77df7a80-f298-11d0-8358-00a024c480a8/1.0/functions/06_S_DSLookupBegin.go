package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSLookupBeginRequest carries the [in] parameters of S_DSLookupBegin.
type s_DSLookupBeginRequest struct {
	PwcsContext  *ndr.WSTR             `ndr:"unique"`
	PRestriction *msmqds.MQRESTRICTION `ndr:"unique"`
	PColumns     msmqds.MQCOLUMNSET
	PSort        *msmqds.MQSORTSET `ndr:"unique"`
	PhServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE
}

func (*s_DSLookupBeginRequest) Opnum() uint16 { return dscomm.OpnumS_DSLookupBegin }

// s_DSLookupBeginResponse carries the [out] parameters and return value of S_DSLookupBegin.
type s_DSLookupBeginResponse struct {
	PHandle msmqds.PPCONTEXT_HANDLE_TYPE
	Status  ndr.DWORD `ndr:"retval"`
}

// S_DSLookupBegin calls S_DSLookupBegin (opnum 6) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSLookupBegin(rpc ndr.Invoker, pwcsContext *ndr.WSTR, pRestriction *msmqds.MQRESTRICTION, pColumns msmqds.MQCOLUMNSET, pSort *msmqds.MQSORTSET, phServerAuth msmqds.PCONTEXT_HANDLE_SERVER_AUTH_TYPE) (PHandle msmqds.PPCONTEXT_HANDLE_TYPE, err error) {
	req := &s_DSLookupBeginRequest{
		PwcsContext:  pwcsContext,
		PRestriction: pRestriction,
		PColumns:     pColumns,
		PSort:        pSort,
		PhServerAuth: phServerAuth,
	}
	var resp s_DSLookupBeginResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSLookupBegin: %w", err)
		return
	}
	PHandle = resp.PHandle
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSLookupBegin failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}
