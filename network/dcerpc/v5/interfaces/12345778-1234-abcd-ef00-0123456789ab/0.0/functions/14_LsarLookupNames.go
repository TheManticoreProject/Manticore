package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarLookupNamesRequest is the [in]/[in,out] parameter set of LsarLookupNames: a policy
// handle, the count of names, a conformant array of names to translate, the [in,out]
// TranslatedSids (a single, inline value), the lookup level, and the [in,out] mapped
// count.
type lsarLookupNamesRequest struct {
	PolicyHandle   structures.LSAPR_HANDLE
	Count          ndr.DWORD
	Names          []dtyp.RPC_UNICODE_STRING `ndr:"conformant,size_is=Count"`
	TranslatedSids structures.LSAPR_TRANSLATED_SIDS
	LookupLevel    structures.LSAP_LOOKUP_LEVEL
	MappedCount    ndr.DWORD
}

func (*lsarLookupNamesRequest) Opnum() uint16 { return lsarpc.OpnumLsarLookupNames }

// lsarLookupNamesResponse is the reply: the [out] referenced domains (a double pointer),
// the [in,out] translated SIDs, the [in,out] mapped count, and the NTSTATUS.
type lsarLookupNamesResponse struct {
	ReferencedDomains *structures.LSAPR_REFERENCED_DOMAIN_LIST `ndr:"unique"`
	TranslatedSids    structures.LSAPR_TRANSLATED_SIDS
	MappedCount       ndr.DWORD
	Status            ndr.DWORD
}

// LsarLookupNames calls LsarLookupNames (opnum 14) to translate a set of account names
// into SIDs ([MS-LSAT] 3.1.4.5). The server may return STATUS_SOME_NOT_MAPPED or
// STATUS_NONE_MAPPED when not all names resolved; in those cases the (partial) results
// are still returned without error, so callers should inspect the translated SIDs and
// mapped count. Any other non-success status is reported as an error.
func LsarLookupNames(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, names []dtyp.RPC_UNICODE_STRING, lookupLevel structures.LSAP_LOOKUP_LEVEL) (*structures.LSAPR_REFERENCED_DOMAIN_LIST, structures.LSAPR_TRANSLATED_SIDS, uint32, error) {
	req := &lsarLookupNamesRequest{
		PolicyHandle: policyHandle,
		Count:        ndr.DWORD(len(names)),
		Names:        names,
		LookupLevel:  lookupLevel,
	}
	var resp lsarLookupNamesResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, structures.LSAPR_TRANSLATED_SIDS{}, 0, fmt.Errorf("LsarLookupNames: %w", err)
	}
	status := uint32(resp.Status)
	switch status {
	case lsarpc.StatusSuccess, lsarpc.StatusSomeNotMapped, lsarpc.StatusNoneMapped:
		return resp.ReferencedDomains, resp.TranslatedSids, uint32(resp.MappedCount), nil
	default:
		return resp.ReferencedDomains, resp.TranslatedSids, uint32(resp.MappedCount), fmt.Errorf("LsarLookupNames failed: %s", lsarpc.StatusString(status))
	}
}
