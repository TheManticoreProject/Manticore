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

// samrQueryInformationDomainRequest is the [in] parameter set of SamrQueryInformationDomain:
// a domain handle and the information class selecting the union arm to return.
type samrQueryInformationDomainRequest struct {
	DomainHandle           mssamr.SAMPR_HANDLE
	DomainInformationClass mssamr.DOMAIN_INFORMATION_CLASS `ndr:"enum"`
}

func (*samrQueryInformationDomainRequest) Opnum() uint16 {
	return samr.OpnumSamrQueryInformationDomain
}

// samrQueryInformationDomainResponse is the reply: the [out, switch_is] double-pointer
// SAMPR_DOMAIN_INFO_BUFFER union (modeled [unique]) and the NTSTATUS.
type samrQueryInformationDomainResponse struct {
	Buffer *mssamr.SAMPR_DOMAIN_INFO_BUFFER `ndr:"unique"`
	Status ndr.DWORD                        `ndr:"retval"`
}

// SamrQueryInformationDomain calls SamrQueryInformationDomain (opnum 8), retrieving domain
// attributes selected by class ([MS-SAMR] 3.1.5.5.2). The returned union carries its own Tag.
func SamrQueryInformationDomain(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, class uint16) (*mssamr.SAMPR_DOMAIN_INFO_BUFFER, error) {
	req := &samrQueryInformationDomainRequest{
		DomainHandle:           domainHandle,
		DomainInformationClass: mssamr.DOMAIN_INFORMATION_CLASS(class),
	}
	var resp samrQueryInformationDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrQueryInformationDomain: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Buffer, fmt.Errorf("SamrQueryInformationDomain failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Buffer, nil
}
