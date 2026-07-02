package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarLookupNames2Request is the [in]/[in,out] parameter set of LsarLookupNames2: like
// LsarLookupNames but TranslatedSids is the EX form and the request carries the trailing
// LookupOptions and ClientRevision fields.
type lsarLookupNames2Request struct {
	PolicyHandle   mslsad.LSAPR_HANDLE
	Count          ndr.DWORD
	Names          []dtyp.RPC_UNICODE_STRING `ndr:"ref,size_is=Count"`
	TranslatedSids mslsad.LSAPR_TRANSLATED_SIDS_EX
	LookupLevel    mslsad.LSAP_LOOKUP_LEVEL `ndr:"enum"`
	MappedCount    ndr.DWORD
	LookupOptions  ndr.DWORD
	ClientRevision ndr.DWORD
}

func (*lsarLookupNames2Request) Opnum() uint16 { return lsarpc.OpnumLsarLookupNames2 }

// lsarLookupNames2Response is the reply: the [out] referenced domains (a double pointer),
// the [in,out] translated SIDs (EX form), the [in,out] mapped count, and the NTSTATUS.
type lsarLookupNames2Response struct {
	ReferencedDomains *mslsad.LSAPR_REFERENCED_DOMAIN_LIST `ndr:"unique"`
	TranslatedSids    mslsad.LSAPR_TRANSLATED_SIDS_EX
	MappedCount       ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// LsarLookupNames2 calls LsarLookupNames2 (opnum 58) to translate a set of account names
// into SIDs with extended results ([MS-LSAT] 3.1.4.6). The server may return
// STATUS_SOME_NOT_MAPPED or STATUS_NONE_MAPPED when not all names resolved; in those cases
// the (partial) results are still returned without error. Any other non-success status is
// reported as an error.
func LsarLookupNames2(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, names []dtyp.RPC_UNICODE_STRING, lookupLevel mslsad.LSAP_LOOKUP_LEVEL, lookupOptions uint32, clientRevision uint32) (*mslsad.LSAPR_REFERENCED_DOMAIN_LIST, mslsad.LSAPR_TRANSLATED_SIDS_EX, uint32, error) {
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
		return nil, mslsad.LSAPR_TRANSLATED_SIDS_EX{}, 0, fmt.Errorf("LsarLookupNames2: %w", err)
	}
	status := uint32(resp.Status)
	switch status {
	case lsarpc.StatusSuccess, lsarpc.StatusSomeNotMapped, lsarpc.StatusNoneMapped:
		return resp.ReferencedDomains, resp.TranslatedSids, uint32(resp.MappedCount), nil
	default:
		return resp.ReferencedDomains, resp.TranslatedSids, uint32(resp.MappedCount), fmt.Errorf("LsarLookupNames2 failed: %s", lsarpc.StatusString(status))
	}
}
