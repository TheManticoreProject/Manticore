package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarLookupNames2Request is the [in]/[in,out] parameter set of LsarLookupNames2: like
// LsarLookupNames but TranslatedSids is the EX form and the request carries the trailing
// LookupOptions and ClientRevision fields.
type lsarLookupNames2Request struct {
	PolicyHandle   structures.LSAPR_HANDLE
	Count          ndr.DWORD
	Names          []dtyp.RPC_UNICODE_STRING `ndr:"ref,size_is=Count"`
	TranslatedSids structures.LSAPR_TRANSLATED_SIDS_EX
	LookupLevel    structures.LSAP_LOOKUP_LEVEL
	MappedCount    ndr.DWORD
	LookupOptions  ndr.DWORD
	ClientRevision ndr.DWORD
}

func (*lsarLookupNames2Request) Opnum() uint16 { return lsarpc.OpnumLsarLookupNames2 }

// lsarLookupNames2Response is the reply: the [out] referenced domains (a double pointer),
// the [in,out] translated SIDs (EX form), the [in,out] mapped count, and the NTSTATUS.
type lsarLookupNames2Response struct {
	ReferencedDomains *structures.LSAPR_REFERENCED_DOMAIN_LIST `ndr:"unique"`
	TranslatedSids    structures.LSAPR_TRANSLATED_SIDS_EX
	MappedCount       ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// LsarLookupNames2 calls LsarLookupNames2 (opnum 58) to translate a set of account names
// into SIDs with extended results ([MS-LSAT] 3.1.4.6). The server may return
// STATUS_SOME_NOT_MAPPED or STATUS_NONE_MAPPED when not all names resolved; in those cases
// the (partial) results are still returned without error. Any other non-success status is
// reported as an error.
func LsarLookupNames2(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, names []dtyp.RPC_UNICODE_STRING, lookupLevel structures.LSAP_LOOKUP_LEVEL, lookupOptions uint32, clientRevision uint32) (*structures.LSAPR_REFERENCED_DOMAIN_LIST, structures.LSAPR_TRANSLATED_SIDS_EX, uint32, error) {
	req := &lsarLookupNames2Request{
		PolicyHandle:   policyHandle,
		Count:          ndr.DWORD(len(names)),
		Names:          names,
		LookupLevel:    lookupLevel,
		LookupOptions:  ndr.DWORD(lookupOptions),
		ClientRevision: ndr.DWORD(clientRevision),
	}
	var resp lsarLookupNames2Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, structures.LSAPR_TRANSLATED_SIDS_EX{}, 0, fmt.Errorf("LsarLookupNames2: %w", err)
	}
	status := uint32(resp.Status)
	switch status {
	case lsarpc.StatusSuccess, lsarpc.StatusSomeNotMapped, lsarpc.StatusNoneMapped:
		return resp.ReferencedDomains, resp.TranslatedSids, uint32(resp.MappedCount), nil
	default:
		return resp.ReferencedDomains, resp.TranslatedSids, uint32(resp.MappedCount), fmt.Errorf("LsarLookupNames2 failed: %s", lsarpc.StatusString(status))
	}
}
