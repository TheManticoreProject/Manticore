package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrLookupNamesInDomainRequest is the [in] parameter set of SamrLookupNamesInDomain: a
// domain handle, the Count of names, and the names themselves.
//
// Names is the IDL's [in, size_is(1000), length_is(Count)] RPC_UNICODE_STRING Names[*],
// modeled as a [ref] pointer to a conformant array (ndr:"ref,size_is=Count"), matching the
// live-validated lsarpc LsarLookupNames pattern. A bare conformant array field hoists its
// maximum_count to the front of the request — ahead of the context handle — which the
// server rejects as nca_s_fault_context_mismatch; the [ref] form emits no referent id and
// defers the array (with its count) to after the handle. Live-validated against Windows XP.
type samrLookupNamesInDomainRequest struct {
	DomainHandle structures.SAMPR_HANDLE
	Count        ndr.DWORD
	Names        []dtyp.RPC_UNICODE_STRING `ndr:"ref,size_is=1000,varying"`
}

func (*samrLookupNamesInDomainRequest) Opnum() uint16 { return samr.OpnumSamrLookupNamesInDomain }

// samrLookupNamesInDomainResponse is the reply: the [out] [ref] RID array and Use array
// (single pointer containers, inline) and the NTSTATUS.
type samrLookupNamesInDomainResponse struct {
	RelativeIds structures.SAMPR_ULONG_ARRAY
	Use         structures.SAMPR_ULONG_ARRAY
	Status      ndr.DWORD `ndr:"retval"`
}

// SamrLookupNamesInDomain calls SamrLookupNamesInDomain (opnum 17), translating account
// names into RIDs and their SID_NAME_USE classifications ([MS-SAMR] 3.1.5.11.2).
func SamrLookupNamesInDomain(rpc *client.Client, domainHandle structures.SAMPR_HANDLE, names []string) (structures.SAMPR_ULONG_ARRAY, structures.SAMPR_ULONG_ARRAY, error) {
	wnames := make([]dtyp.RPC_UNICODE_STRING, len(names))
	for i, n := range names {
		wnames[i] = dtyp.NewUnicodeString(n)
	}
	req := &samrLookupNamesInDomainRequest{
		DomainHandle: domainHandle,
		Count:        ndr.DWORD(len(names)),
		Names:        wnames,
	}
	var resp samrLookupNamesInDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_ULONG_ARRAY{}, structures.SAMPR_ULONG_ARRAY{}, fmt.Errorf("SamrLookupNamesInDomain: %w", err)
	}
	status := uint32(resp.Status)
	if status != samr.StatusSuccess && status != samr.StatusSomeNotMapped && status != samr.StatusNoneMapped {
		return resp.RelativeIds, resp.Use, fmt.Errorf("SamrLookupNamesInDomain failed: %s", samr.StatusString(status))
	}
	return resp.RelativeIds, resp.Use, nil
}
