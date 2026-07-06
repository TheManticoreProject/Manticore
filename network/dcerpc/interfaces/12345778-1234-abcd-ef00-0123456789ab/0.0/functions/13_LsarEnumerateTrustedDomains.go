package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarEnumerateTrustedDomainsRequest is the [in]/[in,out] parameter set of
// LsarEnumerateTrustedDomains: an open policy handle, the [in,out] enumeration context
// (resume handle), and the preferred maximum byte length of the returned data.
type lsarEnumerateTrustedDomainsRequest struct {
	PolicyHandle          mslsad.LSAPR_HANDLE
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
	EnumerationBuffer  mslsad.LSAPR_TRUSTED_ENUM_BUFFER
	Status             ndr.DWORD `ndr:"retval"`
}

// LsarEnumerateTrustedDomains calls LsarEnumerateTrustedDomains (opnum 13), returning a
// page of trusted domains together with the updated enumeration context to resume from.
// The server returns STATUS_SUCCESS or STATUS_MORE_ENTRIES while entries remain and
// STATUS_NO_MORE_ENTRIES once the enumeration is exhausted; all three are treated as
// success here, and the caller continues until STATUS_NO_MORE_ENTRIES.
func LsarEnumerateTrustedDomains(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, enumerationContext uint32, preferedMaximumLength uint32) (uint32, mslsad.LSAPR_TRUSTED_ENUM_BUFFER, error) {
	req := &lsarEnumerateTrustedDomainsRequest{
		PolicyHandle:          policyHandle,
		EnumerationContext:    ndr.DWORD(enumerationContext),
		PreferedMaximumLength: ndr.DWORD(preferedMaximumLength),
	}
	var resp lsarEnumerateTrustedDomainsResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return enumerationContext, mslsad.LSAPR_TRUSTED_ENUM_BUFFER{}, fmt.Errorf("LsarEnumerateTrustedDomains: %w", err)
	}
	switch uint32(resp.Status) {
	case lsarpc.StatusSuccess, lsarpc.StatusMoreEntries, lsarpc.StatusNoMoreEntries:
		return uint32(resp.EnumerationContext), resp.EnumerationBuffer, nil
	default:
		return uint32(resp.EnumerationContext), resp.EnumerationBuffer, fmt.Errorf("LsarEnumerateTrustedDomains failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
}
