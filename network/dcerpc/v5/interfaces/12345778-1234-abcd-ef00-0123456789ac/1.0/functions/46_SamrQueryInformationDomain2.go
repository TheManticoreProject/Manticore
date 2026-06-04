package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrQueryInformationDomain2Request is the [in] parameter set of
// SamrQueryInformationDomain2 (identical shape to opnum 8): a domain handle and the
// information class selecting the union arm to return.
type samrQueryInformationDomain2Request struct {
	DomainHandle           structures.SAMPR_HANDLE
	DomainInformationClass structures.DOMAIN_INFORMATION_CLASS
}

func (*samrQueryInformationDomain2Request) Opnum() uint16 {
	return samr.OpnumSamrQueryInformationDomain2
}

// samrQueryInformationDomain2Response is the reply: the [out, switch_is] double-pointer
// SAMPR_DOMAIN_INFO_BUFFER union (modeled [unique]) and the NTSTATUS.
type samrQueryInformationDomain2Response struct {
	Buffer *structures.SAMPR_DOMAIN_INFO_BUFFER `ndr:"unique"`
	Status ndr.DWORD                            `ndr:"retval"`
}

// SamrQueryInformationDomain2 calls SamrQueryInformationDomain2 (opnum 46), retrieving domain
// attributes selected by class ([MS-SAMR] 3.1.5.5.1). The returned union carries its own Tag.
func SamrQueryInformationDomain2(rpc *client.Client, domainHandle structures.SAMPR_HANDLE, class uint16) (*structures.SAMPR_DOMAIN_INFO_BUFFER, error) {
	req := &samrQueryInformationDomain2Request{
		DomainHandle:           domainHandle,
		DomainInformationClass: structures.DOMAIN_INFORMATION_CLASS(class),
	}
	var resp samrQueryInformationDomain2Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrQueryInformationDomain2: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Buffer, fmt.Errorf("SamrQueryInformationDomain2 failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Buffer, nil
}
