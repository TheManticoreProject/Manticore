package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarLookupSidsRequest is the [in]/[in,out] parameter set of LsarLookupSids: a policy
// handle, the inline SidEnumBuffer (a single [ref] value), the [in,out] TranslatedNames
// (a single, inline value), the lookup level, and the [in,out] mapped count.
type lsarLookupSidsRequest struct {
	PolicyHandle    structures.LSAPR_HANDLE
	SidEnumBuffer   structures.LSAPR_SID_ENUM_BUFFER
	TranslatedNames structures.LSAPR_TRANSLATED_NAMES
	LookupLevel     structures.LSAP_LOOKUP_LEVEL
	MappedCount     ndr.DWORD
}

func (*lsarLookupSidsRequest) Opnum() uint16 { return lsarpc.OpnumLsarLookupSids }

// lsarLookupSidsResponse is the reply: the [out] referenced domains (a double pointer),
// the [in,out] translated names, the [in,out] mapped count, and the NTSTATUS.
type lsarLookupSidsResponse struct {
	ReferencedDomains *structures.LSAPR_REFERENCED_DOMAIN_LIST `ndr:"unique"`
	TranslatedNames   structures.LSAPR_TRANSLATED_NAMES
	MappedCount       ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// LsarLookupSids calls LsarLookupSids (opnum 15) to translate a set of SIDs into account
// names ([MS-LSAT] 3.1.4.7). The server may return STATUS_SOME_NOT_MAPPED or
// STATUS_NONE_MAPPED when not all SIDs resolved; in those cases the (partial) results are
// still returned without error, so callers should inspect the translated names and mapped
// count. Any other non-success status is reported as an error.
func LsarLookupSids(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, sidEnumBuffer structures.LSAPR_SID_ENUM_BUFFER, lookupLevel structures.LSAP_LOOKUP_LEVEL) (*structures.LSAPR_REFERENCED_DOMAIN_LIST, structures.LSAPR_TRANSLATED_NAMES, uint32, error) {
	req := &lsarLookupSidsRequest{
		PolicyHandle:  policyHandle,
		SidEnumBuffer: sidEnumBuffer,
		LookupLevel:   lookupLevel,
	}
	var resp lsarLookupSidsResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, structures.LSAPR_TRANSLATED_NAMES{}, 0, fmt.Errorf("LsarLookupSids: %w", err)
	}
	status := uint32(resp.Status)
	switch status {
	case lsarpc.StatusSuccess, lsarpc.StatusSomeNotMapped, lsarpc.StatusNoneMapped:
		return resp.ReferencedDomains, resp.TranslatedNames, uint32(resp.MappedCount), nil
	default:
		return resp.ReferencedDomains, resp.TranslatedNames, uint32(resp.MappedCount), fmt.Errorf("LsarLookupSids failed: %s", lsarpc.StatusString(status))
	}
}
