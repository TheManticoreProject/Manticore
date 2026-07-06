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

// lsarOpenTrustedDomainByNameRequest is the [in] parameter set of
// LsarOpenTrustedDomainByName: an open policy handle, the [unique] name of the trusted
// domain to open, and the desired access mask for the returned handle.
type lsarOpenTrustedDomainByNameRequest struct {
	PolicyHandle      mslsad.LSAPR_HANDLE
	TrustedDomainName msdtyp.RPC_UNICODE_STRING
	DesiredAccess     ndr.DWORD
}

func (*lsarOpenTrustedDomainByNameRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarOpenTrustedDomainByName
}

// LsarOpenTrustedDomainByName calls LsarOpenTrustedDomainByName (opnum 55), opening the
// trusted domain object identified by name and returning a handle to it.
func LsarOpenTrustedDomainByName(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, trustedDomainName string, desiredAccess uint32) (mslsad.LSAPR_HANDLE, error) {
	name := msdtyp.NewUnicodeString(trustedDomainName)
	req := &lsarOpenTrustedDomainByNameRequest{
		PolicyHandle:      policyHandle,
		TrustedDomainName: name,
		DesiredAccess:     ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslsad.LSAPR_HANDLE{}, fmt.Errorf("LsarOpenTrustedDomainByName: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarOpenTrustedDomainByName failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
