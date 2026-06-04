package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrQueryInformationDomainRequest is the [in] parameter set of SamrQueryInformationDomain:
// a domain handle and the information class selecting the union arm to return.
type samrQueryInformationDomainRequest struct {
	DomainHandle           structures.SAMPR_HANDLE
	DomainInformationClass structures.DOMAIN_INFORMATION_CLASS
}

func (*samrQueryInformationDomainRequest) Opnum() uint16 {
	return samr.OpnumSamrQueryInformationDomain
}

// samrQueryInformationDomainResponse is the reply: the [out, switch_is] double-pointer
// SAMPR_DOMAIN_INFO_BUFFER union (modeled [unique]) and the NTSTATUS.
type samrQueryInformationDomainResponse struct {
	Buffer *structures.SAMPR_DOMAIN_INFO_BUFFER `ndr:"unique"`
	Status ndr.DWORD                            `ndr:"retval"`
}

// SamrQueryInformationDomain calls SamrQueryInformationDomain (opnum 8), retrieving domain
// attributes selected by class ([MS-SAMR] 3.1.5.5.2). The returned union carries its own Tag.
func SamrQueryInformationDomain(rpc *client.Client, domainHandle structures.SAMPR_HANDLE, class uint16) (*structures.SAMPR_DOMAIN_INFO_BUFFER, error) {
	req := &samrQueryInformationDomainRequest{
		DomainHandle:           domainHandle,
		DomainInformationClass: structures.DOMAIN_INFORMATION_CLASS(class),
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
