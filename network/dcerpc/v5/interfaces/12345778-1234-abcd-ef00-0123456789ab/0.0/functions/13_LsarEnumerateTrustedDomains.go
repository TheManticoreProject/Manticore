package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarEnumerateTrustedDomainsRequest is the [in]/[in,out] parameter set of
// LsarEnumerateTrustedDomains: an open policy handle, the [in,out] enumeration context
// (resume handle), and the preferred maximum byte length of the returned data.
type lsarEnumerateTrustedDomainsRequest struct {
	PolicyHandle          structures.LSAPR_HANDLE
	EnumerationContext    ndr.DWORD
	PreferedMaximumLength ndr.DWORD
}

func (*lsarEnumerateTrustedDomainsRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarEnumerateTrustedDomains
}

// lsarEnumerateTrustedDomainsResponse is the reply: the [in,out] enumeration context, the
// [out] enumeration buffer (a top-level [ref] struct, so it is inlined), and the NTSTATUS
// return value.
type lsarEnumerateTrustedDomainsResponse struct {
	EnumerationContext ndr.DWORD
	EnumerationBuffer  structures.LSAPR_TRUSTED_ENUM_BUFFER
	Status             ndr.DWORD `ndr:"retval"`
}

// LsarEnumerateTrustedDomains calls LsarEnumerateTrustedDomains (opnum 13), returning a
// page of trusted domains together with the updated enumeration context to resume from.
// The server returns STATUS_SUCCESS or STATUS_MORE_ENTRIES while entries remain and
// STATUS_NO_MORE_ENTRIES once the enumeration is exhausted; all three are treated as
// success here, and the caller continues until STATUS_NO_MORE_ENTRIES.
func LsarEnumerateTrustedDomains(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, enumerationContext uint32, preferedMaximumLength uint32) (uint32, structures.LSAPR_TRUSTED_ENUM_BUFFER, error) {
	req := &lsarEnumerateTrustedDomainsRequest{
		PolicyHandle:          policyHandle,
		EnumerationContext:    ndr.DWORD(enumerationContext),
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
	}
	var resp lsarEnumerateTrustedDomainsResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return enumerationContext, structures.LSAPR_TRUSTED_ENUM_BUFFER{}, fmt.Errorf("LsarEnumerateTrustedDomains: %w", err)
	}
	switch uint32(resp.Status) {
	case lsarpc.StatusSuccess, lsarpc.StatusMoreEntries, lsarpc.StatusNoMoreEntries:
		return uint32(resp.EnumerationContext), resp.EnumerationBuffer, nil
	default:
		return uint32(resp.EnumerationContext), resp.EnumerationBuffer, fmt.Errorf("LsarEnumerateTrustedDomains failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
}
