package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrLookupIdsInDomainRequest is the [in] parameter set of SamrLookupIdsInDomain: a domain
// handle, the Count of RIDs, and the RIDs themselves.
//
// RelativeIds is the IDL's [in, size_is(1000), length_is(Count)] unsigned long*
// RelativeIds, modeled as a [ref] pointer to a conformant array (ndr:"ref,size_is=Count")
// like SamrLookupNamesInDomain: a bare conformant array field would hoist its maximum_count
// ahead of the context handle (nca_s_fault_context_mismatch). The [ref] form defers the
// array past the handle.
type samrLookupIdsInDomainRequest struct {
	DomainHandle mssamr.SAMPR_HANDLE
	Count        ndr.DWORD
	RelativeIds  []ndr.DWORD `ndr:"ref,size_is=1000,varying"`
}

func (*samrLookupIdsInDomainRequest) Opnum() uint16 { return samr.OpnumSamrLookupIdsInDomain }

// samrLookupIdsInDomainResponse is the reply: the [out] [ref] returned-string array of names
// and the Use array (single pointer containers, inline) and the NTSTATUS.
type samrLookupIdsInDomainResponse struct {
	Names  mssamr.SAMPR_RETURNED_USTRING_ARRAY
	Use    mssamr.SAMPR_ULONG_ARRAY
	Status ndr.DWORD `ndr:"retval"`
}

// SamrLookupIdsInDomain calls SamrLookupIdsInDomain (opnum 18), translating RIDs into
// account names and their SID_NAME_USE classifications ([MS-SAMR] 3.1.5.11.1).
func SamrLookupIdsInDomain(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, relativeIds []uint32) (mssamr.SAMPR_RETURNED_USTRING_ARRAY, mssamr.SAMPR_ULONG_ARRAY, error) {
	rids := make([]ndr.DWORD, len(relativeIds))
	for i, r := range relativeIds {
		rids[i] = ndr.DWORD(r)
	}
	req := &samrLookupIdsInDomainRequest{
		DomainHandle: domainHandle,
		Count:        ndr.DWORD(len(relativeIds)),
		RelativeIds:  rids,
	}
	var resp samrLookupIdsInDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.SAMPR_RETURNED_USTRING_ARRAY{}, mssamr.SAMPR_ULONG_ARRAY{}, fmt.Errorf("SamrLookupIdsInDomain: %w", err)
	}
	status := uint32(resp.Status)
	if status != samr.StatusSuccess && status != samr.StatusSomeNotMapped && status != samr.StatusNoneMapped {
		return resp.Names, resp.Use, fmt.Errorf("SamrLookupIdsInDomain failed: %s", samr.StatusString(status))
	}
	return resp.Names, resp.Use, nil
}
