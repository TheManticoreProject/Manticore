package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrEnumerateAliasesInDomainRequest carries the [in] domain handle, the [in,out]
// enumeration context, and the preferred maximum response length.
type samrEnumerateAliasesInDomainRequest struct {
	DomainHandle          mssamr.SAMPR_HANDLE
	EnumerationContext    ndr.DWORD
	PreferedMaximumLength ndr.DWORD
}

func (*samrEnumerateAliasesInDomainRequest) Opnum() uint16 {
	return samr.OpnumSamrEnumerateAliasesInDomain
}

// samrEnumerateAliasesInDomainResponse is the reply: the updated [in,out] enumeration
// context, the [unique] enumeration buffer, the count returned, and the NTSTATUS.
type samrEnumerateAliasesInDomainResponse struct {
	EnumerationContext ndr.DWORD
	Buffer             *mssamr.SAMPR_ENUMERATION_BUFFER `ndr:"unique"`
	CountReturned      ndr.DWORD
	Status             ndr.DWORD `ndr:"retval"`
}

// SamrEnumerateAliasesInDomain calls SamrEnumerateAliasesInDomain (opnum 15), enumerating
// the aliases of the domain ([MS-SAMR] 3.1.5.2.4). STATUS_MORE_ENTRIES is a success-style
// continuation status, not an error.
func SamrEnumerateAliasesInDomain(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, enumerationContext uint32, preferedMaximumLength uint32) (uint32, *mssamr.SAMPR_ENUMERATION_BUFFER, uint32, error) {
	req := &samrEnumerateAliasesInDomainRequest{
		DomainHandle:          domainHandle,
		EnumerationContext:    ndr.DWORD(enumerationContext),
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
	}
	var resp samrEnumerateAliasesInDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, nil, 0, fmt.Errorf("SamrEnumerateAliasesInDomain: %w", err)
	}
	status := uint32(resp.Status)
	if status != samr.StatusSuccess && status != samr.StatusMoreEntries {
		return uint32(resp.EnumerationContext), resp.Buffer, uint32(resp.CountReturned), fmt.Errorf("SamrEnumerateAliasesInDomain failed: %s", samr.StatusString(status))
	}
	return uint32(resp.EnumerationContext), resp.Buffer, uint32(resp.CountReturned), nil
}
