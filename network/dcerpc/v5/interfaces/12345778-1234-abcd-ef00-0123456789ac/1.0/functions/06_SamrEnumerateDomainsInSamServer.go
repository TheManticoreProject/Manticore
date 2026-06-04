package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrEnumerateDomainsInSamServerRequest carries the [in] parameters: the server
// handle, the [in,out] enumeration context (a cookie that resumes a prior call),
// and the preferred maximum byte length of the returned buffer.
type samrEnumerateDomainsInSamServerRequest struct {
	ServerHandle          structures.SAMPR_HANDLE
	EnumerationContext    ndr.DWORD
	PreferedMaximumLength ndr.DWORD
}

func (*samrEnumerateDomainsInSamServerRequest) Opnum() uint16 {
	return samr.OpnumSamrEnumerateDomainsInSamServer
}

// samrEnumerateDomainsInSamServerResponse carries the [out]/[in,out] parameters:
// the updated enumeration context, the [unique] enumeration buffer, the count of
// entries returned, and the NTSTATUS.
type samrEnumerateDomainsInSamServerResponse struct {
	EnumerationContext ndr.DWORD
	Buffer             *structures.SAMPR_ENUMERATION_BUFFER `ndr:"unique"`
	CountReturned      ndr.DWORD
	Status             ndr.DWORD `ndr:"retval"`
}

// SamrEnumerateDomainsInSamServer calls SamrEnumerateDomainsInSamServer (opnum 6),
// listing the domains hosted by the server ([MS-SAMR] 3.1.5.2.1). A return of
// STATUS_MORE_ENTRIES indicates further entries remain; resume with the returned
// enumeration context.
func SamrEnumerateDomainsInSamServer(rpc *client.Client, handle structures.SAMPR_HANDLE, enumerationContext uint32, preferedMaximumLength uint32) (uint32, *structures.SAMPR_ENUMERATION_BUFFER, uint32, error) {
	req := &samrEnumerateDomainsInSamServerRequest{
		ServerHandle:          handle,
		EnumerationContext:    ndr.DWORD(enumerationContext),
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
	}
	var resp samrEnumerateDomainsInSamServerResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, nil, 0, fmt.Errorf("SamrEnumerateDomainsInSamServer: %w", err)
	}
	status := uint32(resp.Status)
	if status != samr.StatusSuccess && status != samr.StatusMoreEntries {
		return uint32(resp.EnumerationContext), resp.Buffer, uint32(resp.CountReturned), fmt.Errorf("SamrEnumerateDomainsInSamServer failed: %s", samr.StatusString(status))
	}
	return uint32(resp.EnumerationContext), resp.Buffer, uint32(resp.CountReturned), nil
}
