package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrSetInformationDomainRequest is the [in] parameter set of SamrSetInformationDomain: a
// domain handle, the information class, and the [in, switch_is] single-pointer union
// (inline) carrying the new values.
type samrSetInformationDomainRequest struct {
	DomainHandle           mssamr.SAMPR_HANDLE
	DomainInformationClass mssamr.DOMAIN_INFORMATION_CLASS `ndr:"enum"`
	DomainInformation      mssamr.SAMPR_DOMAIN_INFO_BUFFER
}

func (*samrSetInformationDomainRequest) Opnum() uint16 {
	return samr.OpnumSamrSetInformationDomain
}

// SamrSetInformationDomain calls SamrSetInformationDomain (opnum 9), updating domain
// attributes selected by class ([MS-SAMR] 3.1.5.4.1). The information union's Tag is set
// to match the class.
func SamrSetInformationDomain(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, class uint16, info mssamr.SAMPR_DOMAIN_INFO_BUFFER) error {
	info.Tag = mssamr.DOMAIN_INFORMATION_CLASS(class)
	req := &samrSetInformationDomainRequest{
		DomainHandle:           domainHandle,
		DomainInformationClass: mssamr.DOMAIN_INFORMATION_CLASS(class),
		DomainInformation:      info,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrSetInformationDomain: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrSetInformationDomain failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
