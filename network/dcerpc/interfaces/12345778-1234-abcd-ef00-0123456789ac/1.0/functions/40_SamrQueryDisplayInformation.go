package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrQueryDisplayInformationRequest is the [in] parameter set of SamrQueryDisplayInformation:
// a domain handle, the display class, the starting index, the requested entry count, and the
// preferred maximum byte length of the returned buffer.
type samrQueryDisplayInformationRequest struct {
	DomainHandle            mssamr.SAMPR_HANDLE
	DisplayInformationClass mssamr.DOMAIN_DISPLAY_INFORMATION `ndr:"enum"`
	Index                   ndr.DWORD
	EntryCount              ndr.DWORD
	PreferredMaximumLength  ndr.DWORD
}

func (*samrQueryDisplayInformationRequest) Opnum() uint16 {
	return samr.OpnumSamrQueryDisplayInformation
}

// samrQueryDisplayInformationResponse is the reply: the total available and total returned
// byte counts, the [out, switch_is] SINGLE-pointer SAMPR_DISPLAY_INFO_BUFFER union (inline,
// [ref] — not [unique], not double), and the NTSTATUS.
type samrQueryDisplayInformationResponse struct {
	TotalAvailable ndr.DWORD
	TotalReturned  ndr.DWORD
	Buffer         mssamr.SAMPR_DISPLAY_INFO_BUFFER
	Status         ndr.DWORD `ndr:"retval"`
}

// SamrQueryDisplayInformation calls SamrQueryDisplayInformation (opnum 40), returning a page
// of account information for display ([MS-SAMR] 3.1.5.3.1). The returned union carries its own
// Tag.
func SamrQueryDisplayInformation(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, class uint16, index, entryCount, preferredMaximumLength uint32) (uint32, uint32, mssamr.SAMPR_DISPLAY_INFO_BUFFER, error) {
	req := &samrQueryDisplayInformationRequest{
		DomainHandle:            domainHandle,
		DisplayInformationClass: mssamr.DOMAIN_DISPLAY_INFORMATION(class),
		Index:                   ndr.DWORD(index),
		EntryCount:              ndr.DWORD(entryCount),
		PreferredMaximumLength:  ndr.DWORD(preferredMaximumLength),
	}
	var resp samrQueryDisplayInformationResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, 0, mssamr.SAMPR_DISPLAY_INFO_BUFFER{}, fmt.Errorf("SamrQueryDisplayInformation: %w", err)
	}
	status := uint32(resp.Status)
	if status != samr.StatusSuccess && status != samr.StatusMoreEntries {
		return uint32(resp.TotalAvailable), uint32(resp.TotalReturned), resp.Buffer, fmt.Errorf("SamrQueryDisplayInformation failed: %s", samr.StatusString(status))
	}
	return uint32(resp.TotalAvailable), uint32(resp.TotalReturned), resp.Buffer, nil
}
