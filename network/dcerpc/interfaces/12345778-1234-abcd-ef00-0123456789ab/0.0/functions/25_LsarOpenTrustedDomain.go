package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarOpenTrustedDomainRequest is the [in] parameter set of LsarOpenTrustedDomain: an open
// policy handle, the [unique] SID of the trusted domain to open, and the desired access
// mask for the returned handle.
type lsarOpenTrustedDomainRequest struct {
	PolicyHandle     mslsad.LSAPR_HANDLE
	TrustedDomainSid *msdtyp.RPC_SID `ndr:"unique"`
	DesiredAccess    ndr.DWORD
}

func (*lsarOpenTrustedDomainRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarOpenTrustedDomain
}

// LsarOpenTrustedDomain calls LsarOpenTrustedDomain (opnum 25), opening the trusted domain
// object identified by its SID and returning a handle to it.
func LsarOpenTrustedDomain(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, trustedDomainSid *msdtyp.RPC_SID, desiredAccess uint32) (mslsad.LSAPR_HANDLE, error) {
	req := &lsarOpenTrustedDomainRequest{
		PolicyHandle:     policyHandle,
		TrustedDomainSid: trustedDomainSid,
		DesiredAccess:    ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_HANDLE{}, fmt.Errorf("LsarOpenTrustedDomain: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarOpenTrustedDomain failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
