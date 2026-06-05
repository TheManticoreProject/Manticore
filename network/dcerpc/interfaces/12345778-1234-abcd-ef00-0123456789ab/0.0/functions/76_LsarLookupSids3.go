package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// lsarLookupSids3Request is the [in]/[in,out] parameter set of LsarLookupSids3. The first
// IDL parameter, [in] handle_t RpcHandle, is the RPC binding handle and is not marshalled,
// so it is omitted here. The request is otherwise the same as LsarLookupSids2 (EX
// translated names plus LookupOptions/ClientRevision).
type lsarLookupSids3Request struct {
	SidEnumBuffer   structures.LSAPR_SID_ENUM_BUFFER
	TranslatedNames structures.LSAPR_TRANSLATED_NAMES_EX
	LookupLevel     structures.LSAP_LOOKUP_LEVEL
	MappedCount     ndr.DWORD
	LookupOptions   ndr.DWORD
	ClientRevision  ndr.DWORD
}

func (*lsarLookupSids3Request) Opnum() uint16 { return lsarpc.OpnumLsarLookupSids3 }

// lsarLookupSids3Response is the reply: the [out] referenced domains (a double pointer),
// the [in,out] translated names (EX form), the [in,out] mapped count, and the NTSTATUS.
type lsarLookupSids3Response struct {
	ReferencedDomains *structures.LSAPR_REFERENCED_DOMAIN_LIST `ndr:"unique"`
	TranslatedNames   structures.LSAPR_TRANSLATED_NAMES_EX
	MappedCount       ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// LsarLookupSids3 calls LsarLookupSids3 (opnum 76) to translate a set of SIDs into account
// names with extended results, without a policy handle ([MS-LSAT] 3.1.4.10). The leading
// handle_t binding handle is supplied implicitly by the RPC runtime and is not part of the
// marshalled request. The server may return STATUS_SOME_NOT_MAPPED or STATUS_NONE_MAPPED
// when not all SIDs resolved; in those cases the (partial) results are still returned
// without error. Any other non-success status is reported as an error.
func LsarLookupSids3(rpc ndr.Invoker, sidEnumBuffer structures.LSAPR_SID_ENUM_BUFFER, lookupLevel structures.LSAP_LOOKUP_LEVEL, lookupOptions uint32, clientRevision uint32) (*structures.LSAPR_REFERENCED_DOMAIN_LIST, structures.LSAPR_TRANSLATED_NAMES_EX, uint32, error) {
	req := &lsarLookupSids3Request{
		SidEnumBuffer:  sidEnumBuffer,
		LookupLevel:    lookupLevel,
		LookupOptions:  ndr.DWORD(lookupOptions),
		ClientRevision: ndr.DWORD(clientRevision),
	}
	var resp lsarLookupSids3Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, structures.LSAPR_TRANSLATED_NAMES_EX{}, 0, fmt.Errorf("LsarLookupSids3: %w", err)
	}
	status := uint32(resp.Status)
	switch status {
	case lsarpc.StatusSuccess, lsarpc.StatusSomeNotMapped, lsarpc.StatusNoneMapped:
		return resp.ReferencedDomains, resp.TranslatedNames, uint32(resp.MappedCount), nil
	default:
		return resp.ReferencedDomains, resp.TranslatedNames, uint32(resp.MappedCount), fmt.Errorf("LsarLookupSids3 failed: %s", lsarpc.StatusString(status))
	}
}
