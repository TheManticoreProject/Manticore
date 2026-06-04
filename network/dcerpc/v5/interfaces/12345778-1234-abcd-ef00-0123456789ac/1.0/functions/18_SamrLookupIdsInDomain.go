package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrLookupIdsInDomainRequest is the [in] parameter set of SamrLookupIdsInDomain: a domain
// handle, the Count of RIDs, and the RIDs themselves.
//
// WIRE RISK: the IDL declares RelativeIds as [in, size_is(1000), length_is(Count)]
// unsigned long *RelativeIds — a fixed maximum (1000) but a varying actual length tied to
// Count. It is modeled here as a conformant,varying slice; the conformance bound on the
// wire may need to be the constant 1000 rather than Count. Verify against a live server.
type samrLookupIdsInDomainRequest struct {
	DomainHandle structures.SAMPR_HANDLE
	Count        ndr.DWORD
	RelativeIds  []ndr.DWORD `ndr:"conformant,varying"`
}

func (*samrLookupIdsInDomainRequest) Opnum() uint16 { return samr.OpnumSamrLookupIdsInDomain }

// samrLookupIdsInDomainResponse is the reply: the [out] [ref] returned-string array of names
// and the Use array (single pointer containers, inline) and the NTSTATUS.
type samrLookupIdsInDomainResponse struct {
	Names  structures.SAMPR_RETURNED_USTRING_ARRAY
	Use    structures.SAMPR_ULONG_ARRAY
	Status ndr.DWORD `ndr:"retval"`
}

// SamrLookupIdsInDomain calls SamrLookupIdsInDomain (opnum 18), translating RIDs into
// account names and their SID_NAME_USE classifications ([MS-SAMR] 3.1.5.11.1).
func SamrLookupIdsInDomain(rpc *client.Client, domainHandle structures.SAMPR_HANDLE, relativeIds []uint32) (structures.SAMPR_RETURNED_USTRING_ARRAY, structures.SAMPR_ULONG_ARRAY, error) {
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
		return structures.SAMPR_RETURNED_USTRING_ARRAY{}, structures.SAMPR_ULONG_ARRAY{}, fmt.Errorf("SamrLookupIdsInDomain: %w", err)
	}
	status := uint32(resp.Status)
	if status != samr.StatusSuccess && status != samr.StatusSomeNotMapped && status != samr.StatusNoneMapped {
		return resp.Names, resp.Use, fmt.Errorf("SamrLookupIdsInDomain failed: %s", samr.StatusString(status))
	}
	return resp.Names, resp.Use, nil
}
