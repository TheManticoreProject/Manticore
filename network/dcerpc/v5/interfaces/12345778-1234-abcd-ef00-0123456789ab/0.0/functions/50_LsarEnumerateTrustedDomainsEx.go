package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarEnumerateTrustedDomainsExRequest is the [in]/[in,out] parameter set of
// LsarEnumerateTrustedDomainsEx: an open policy handle, the [in,out] enumeration context
// (resume handle), and the preferred maximum byte length of the returned data.
type lsarEnumerateTrustedDomainsExRequest struct {
	PolicyHandle          structures.LSAPR_HANDLE
	EnumerationContext    ndr.DWORD
	PreferedMaximumLength ndr.DWORD
}

func (*lsarEnumerateTrustedDomainsExRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarEnumerateTrustedDomainsEx
}

// lsarEnumerateTrustedDomainsExResponse is the reply: the [in,out] enumeration context,
// the [out] enumeration buffer (a top-level [ref] struct, so it is inlined), and the
// NTSTATUS return value.
type lsarEnumerateTrustedDomainsExResponse struct {
	EnumerationContext ndr.DWORD
	EnumerationBuffer  structures.LSAPR_TRUSTED_ENUM_BUFFER_EX
	Status             ndr.DWORD `ndr:"retval"`
}

// LsarEnumerateTrustedDomainsEx calls LsarEnumerateTrustedDomainsEx (opnum 50), returning a
// page of extended trusted-domain information together with the updated enumeration context
// to resume from. The server returns STATUS_SUCCESS or STATUS_MORE_ENTRIES while entries
// remain and STATUS_NO_MORE_ENTRIES once the enumeration is exhausted; all three are
// treated as success here, and the caller continues until STATUS_NO_MORE_ENTRIES.
func LsarEnumerateTrustedDomainsEx(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, enumerationContext uint32, preferedMaximumLength uint32) (uint32, structures.LSAPR_TRUSTED_ENUM_BUFFER_EX, error) {
	req := &lsarEnumerateTrustedDomainsExRequest{
		PolicyHandle:          policyHandle,
		EnumerationContext:    ndr.DWORD(enumerationContext),
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
	}
	var resp lsarEnumerateTrustedDomainsExResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return enumerationContext, structures.LSAPR_TRUSTED_ENUM_BUFFER_EX{}, fmt.Errorf("LsarEnumerateTrustedDomainsEx: %w", err)
	}
	switch uint32(resp.Status) {
	case lsarpc.StatusSuccess, lsarpc.StatusMoreEntries, lsarpc.StatusNoMoreEntries:
		return uint32(resp.EnumerationContext), resp.EnumerationBuffer, nil
	default:
		return uint32(resp.EnumerationContext), resp.EnumerationBuffer, fmt.Errorf("LsarEnumerateTrustedDomainsEx failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
}
