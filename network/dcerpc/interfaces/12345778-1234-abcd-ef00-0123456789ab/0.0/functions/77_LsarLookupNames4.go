package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarLookupNames4Request is the [in]/[in,out] parameter set of LsarLookupNames4. The first
// IDL parameter, [in] handle_t RpcHandle, is the RPC binding handle and is not marshalled,
// so it is omitted here. The request is otherwise the same as LsarLookupNames3 (EX2
// translated SIDs plus LookupOptions/ClientRevision).
type lsarLookupNames4Request struct {
	Count          ndr.DWORD
	Names          []dtyp.RPC_UNICODE_STRING `ndr:"ref,size_is=Count"`
	TranslatedSids mslsad.LSAPR_TRANSLATED_SIDS_EX2
	LookupLevel    mslsad.LSAP_LOOKUP_LEVEL `ndr:"enum"`
	MappedCount    ndr.DWORD
	LookupOptions  ndr.DWORD
	ClientRevision ndr.DWORD
}

func (*lsarLookupNames4Request) Opnum() uint16 { return lsarpc.OpnumLsarLookupNames4 }

// lsarLookupNames4Response is the reply: the [out] referenced domains (a double pointer),
// the [in,out] translated SIDs (EX2 form), the [in,out] mapped count, and the NTSTATUS.
type lsarLookupNames4Response struct {
	ReferencedDomains *mslsad.LSAPR_REFERENCED_DOMAIN_LIST `ndr:"unique"`
	TranslatedSids    mslsad.LSAPR_TRANSLATED_SIDS_EX2
	MappedCount       ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// LsarLookupNames4 calls LsarLookupNames4 (opnum 77) to translate a set of account names
// into SIDs with extended (EX2) results, without a policy handle ([MS-LSAT] 3.1.4.4). The
// leading handle_t binding handle is supplied implicitly by the RPC runtime and is not part
// of the marshalled request. The server may return STATUS_SOME_NOT_MAPPED or
// STATUS_NONE_MAPPED when not all names resolved; in those cases the (partial) results are
// still returned without error. Any other non-success status is reported as an error.
func LsarLookupNames4(rpc ndr.Invoker, names []dtyp.RPC_UNICODE_STRING, lookupLevel mslsad.LSAP_LOOKUP_LEVEL, lookupOptions uint32, clientRevision uint32) (*mslsad.LSAPR_REFERENCED_DOMAIN_LIST, mslsad.LSAPR_TRANSLATED_SIDS_EX2, uint32, error) {
	req := &lsarLookupNames4Request{
		Count:          ndr.DWORD(len(names)),
		Names:          names,
		LookupLevel:    lookupLevel,
		LookupOptions:  ndr.DWORD(lookupOptions),
		ClientRevision: ndr.DWORD(clientRevision),
	}
	var resp lsarLookupNames4Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, mslsad.LSAPR_TRANSLATED_SIDS_EX2{}, 0, fmt.Errorf("LsarLookupNames4: %w", err)
	}
	status := uint32(resp.Status)
	switch status {
	case lsarpc.StatusSuccess, lsarpc.StatusSomeNotMapped, lsarpc.StatusNoneMapped:
		return resp.ReferencedDomains, resp.TranslatedSids, uint32(resp.MappedCount), nil
	default:
		return resp.ReferencedDomains, resp.TranslatedSids, uint32(resp.MappedCount), fmt.Errorf("LsarLookupNames4 failed: %s", lsarpc.StatusString(status))
	}
}
