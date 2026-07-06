package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
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
	DomainHandle mssamr.SAMPR_HANDLE
	Count        ndr.DWORD
	Names        []msdtyp.RPC_UNICODE_STRING `ndr:"ref,size_is=1000,varying"`
}

func (*samrLookupNamesInDomainRequest) Opnum() uint16 { return samr.OpnumSamrLookupNamesInDomain }

// samrLookupNamesInDomainResponse is the reply: the [out] [ref] RID array and Use array
// (single pointer containers, inline) and the NTSTATUS.
type samrLookupNamesInDomainResponse struct {
	RelativeIds mssamr.SAMPR_ULONG_ARRAY
	Use         mssamr.SAMPR_ULONG_ARRAY
	Status      ndr.DWORD `ndr:"retval"`
}

// SamrLookupNamesInDomain calls SamrLookupNamesInDomain (opnum 17), translating account
// names into RIDs and their SID_NAME_USE classifications ([MS-SAMR] 3.1.5.11.2).
func SamrLookupNamesInDomain(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, names []string) (mssamr.SAMPR_ULONG_ARRAY, mssamr.SAMPR_ULONG_ARRAY, error) {
	wnames := make([]msdtyp.RPC_UNICODE_STRING, len(names))
	for i, n := range names {
		wnames[i] = msdtyp.NewUnicodeString(n)
	}
	req := &samrLookupNamesInDomainRequest{
		DomainHandle: domainHandle,
		Count:        ndr.DWORD(len(names)),
		Names:        wnames,
	}
	var resp samrLookupNamesInDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.SAMPR_ULONG_ARRAY{}, mssamr.SAMPR_ULONG_ARRAY{}, fmt.Errorf("SamrLookupNamesInDomain: %w", err)
	}
	status := uint32(resp.Status)
	if status != samr.StatusSuccess && status != samr.StatusSomeNotMapped && status != samr.StatusNoneMapped {
		return resp.RelativeIds, resp.Use, fmt.Errorf("SamrLookupNamesInDomain failed: %s", samr.StatusString(status))
	}
	return resp.RelativeIds, resp.Use, nil
}
