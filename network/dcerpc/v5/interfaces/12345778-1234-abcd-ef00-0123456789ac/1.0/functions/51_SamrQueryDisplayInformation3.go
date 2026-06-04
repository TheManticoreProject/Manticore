package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrQueryDisplayInformation3Request is the [in] parameter set of
// SamrQueryDisplayInformation3 (identical shape to opnum 40): a domain handle, the display
// class, the starting index, the requested entry count, and the preferred maximum byte
// length of the returned buffer.
type samrQueryDisplayInformation3Request struct {
	DomainHandle            structures.SAMPR_HANDLE
	DisplayInformationClass structures.DOMAIN_DISPLAY_INFORMATION
	Index                   ndr.DWORD
	EntryCount              ndr.DWORD
	PreferredMaximumLength  ndr.DWORD
}

func (*samrQueryDisplayInformation3Request) Opnum() uint16 {
	return samr.OpnumSamrQueryDisplayInformation3
}

// samrQueryDisplayInformation3Response is the reply: the total available and total returned
// byte counts, the [out, switch_is] SINGLE-pointer SAMPR_DISPLAY_INFO_BUFFER union (inline,
// [ref] — not [unique], not double), and the NTSTATUS.
type samrQueryDisplayInformation3Response struct {
	TotalAvailable ndr.DWORD
	TotalReturned  ndr.DWORD
	Buffer         structures.SAMPR_DISPLAY_INFO_BUFFER
	Status         ndr.DWORD `ndr:"retval"`
}

// SamrQueryDisplayInformation3 calls SamrQueryDisplayInformation3 (opnum 51), returning a page
// of account information for display ([MS-SAMR] 3.1.5.3.1). The returned union carries its own
// Tag.
func SamrQueryDisplayInformation3(rpc *client.Client, domainHandle structures.SAMPR_HANDLE, class uint16, index, entryCount, preferredMaximumLength uint32) (uint32, uint32, structures.SAMPR_DISPLAY_INFO_BUFFER, error) {
	req := &samrQueryDisplayInformation3Request{
		DomainHandle:            domainHandle,
		DisplayInformationClass: structures.DOMAIN_DISPLAY_INFORMATION(class),
		Index:                   ndr.DWORD(index),
		EntryCount:              ndr.DWORD(entryCount),
		PreferredMaximumLength:  ndr.DWORD(preferredMaximumLength),
	}
	var resp samrQueryDisplayInformation3Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, 0, structures.SAMPR_DISPLAY_INFO_BUFFER{}, fmt.Errorf("SamrQueryDisplayInformation3: %w", err)
	}
	status := uint32(resp.Status)
	if status != samr.StatusSuccess && status != samr.StatusMoreEntries {
		return uint32(resp.TotalAvailable), uint32(resp.TotalReturned), resp.Buffer, fmt.Errorf("SamrQueryDisplayInformation3 failed: %s", samr.StatusString(status))
	}
	return uint32(resp.TotalAvailable), uint32(resp.TotalReturned), resp.Buffer, nil
}
