package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
	mslsat "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsat"
)

// lsarLookupNames3Request is the [in]/[in,out] parameter set of LsarLookupNames3: like
// LsarLookupNames2 but TranslatedSids is the EX2 form.
type lsarLookupNames3Request struct {
	PolicyHandle   mslsad.LSAPR_HANDLE
	Count          ndr.DWORD
	Names          []msdtyp.RPC_UNICODE_STRING `ndr:"ref,size_is=Count"`
	TranslatedSids mslsat.LSAPR_TRANSLATED_SIDS_EX2
	LookupLevel    mslsat.LSAP_LOOKUP_LEVEL `ndr:"enum"`
	MappedCount    ndr.DWORD
	LookupOptions  ndr.DWORD
	ClientRevision ndr.DWORD
}

func (*lsarLookupNames3Request) Opnum() uint16 { return lsarpc.OpnumLsarLookupNames3 }

// lsarLookupNames3Response is the reply: the [out] referenced domains (a double pointer),
// the [in,out] translated SIDs (EX2 form), the [in,out] mapped count, and the NTSTATUS.
type lsarLookupNames3Response struct {
	ReferencedDomains *mslsat.LSAPR_REFERENCED_DOMAIN_LIST `ndr:"unique"`
	TranslatedSids    mslsat.LSAPR_TRANSLATED_SIDS_EX2
	MappedCount       ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// LsarLookupNames3 calls LsarLookupNames3 (opnum 68) to translate a set of account names
// into SIDs with extended (EX2) results ([MS-LSAT] 3.1.4.5 / 3.1.4.6 variants). The server
// may return STATUS_SOME_NOT_MAPPED or STATUS_NONE_MAPPED when not all names resolved; in
// those cases the (partial) results are still returned without error. Any other
// non-success status is reported as an error.
func LsarLookupNames3(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, names []msdtyp.RPC_UNICODE_STRING, lookupLevel mslsat.LSAP_LOOKUP_LEVEL, lookupOptions uint32, clientRevision uint32) (*mslsat.LSAPR_REFERENCED_DOMAIN_LIST, mslsat.LSAPR_TRANSLATED_SIDS_EX2, uint32, error) {
	req := &lsarLookupNames3Request{
		PolicyHandle:   policyHandle,
		Count:          ndr.DWORD(len(names)),
		Names:          names,
		LookupLevel:    lookupLevel,
		LookupOptions:  ndr.DWORD(lookupOptions),
		ClientRevision: ndr.DWORD(clientRevision),
	}
	var resp lsarLookupNames3Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, mslsat.LSAPR_TRANSLATED_SIDS_EX2{}, 0, fmt.Errorf("LsarLookupNames3: %w", err)
	}
	status := uint32(resp.Status)
	switch status {
	case lsarpc.StatusSuccess, lsarpc.StatusSomeNotMapped, lsarpc.StatusNoneMapped:
		return resp.ReferencedDomains, resp.TranslatedSids, uint32(resp.MappedCount), nil
	default:
		return resp.ReferencedDomains, resp.TranslatedSids, uint32(resp.MappedCount), fmt.Errorf("LsarLookupNames3 failed: %s", lsarpc.StatusString(status))
	}
}
