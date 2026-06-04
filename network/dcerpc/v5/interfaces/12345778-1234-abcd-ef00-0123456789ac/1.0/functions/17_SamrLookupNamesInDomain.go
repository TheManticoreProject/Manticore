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
// WIRE RISK: the IDL declares Names as [in, size_is(1000), length_is(Count)]
// RPC_UNICODE_STRING Names[*] — a fixed maximum (1000) but a varying actual length tied to
// Count. It is modeled here as a conformant,varying slice; the conformance bound on the
// wire may need to be the constant 1000 rather than Count. Verify against a live server.
type samrLookupNamesInDomainRequest struct {
	DomainHandle structures.SAMPR_HANDLE
	Count        ndr.DWORD
	Names        []dtyp.RPC_UNICODE_STRING `ndr:"conformant,varying"`
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
