package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarOpenTrustedDomainByNameRequest is the [in] parameter set of
// LsarOpenTrustedDomainByName: an open policy handle, the [unique] name of the trusted
// domain to open, and the desired access mask for the returned handle.
type lsarOpenTrustedDomainByNameRequest struct {
	PolicyHandle      mslsad.LSAPR_HANDLE
	TrustedDomainName dtyp.RPC_UNICODE_STRING
	DesiredAccess     ndr.DWORD
}

func (*lsarOpenTrustedDomainByNameRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarOpenTrustedDomainByName
}

// LsarOpenTrustedDomainByName calls LsarOpenTrustedDomainByName (opnum 55), opening the
// trusted domain object identified by name and returning a handle to it.
func LsarOpenTrustedDomainByName(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, trustedDomainName string, desiredAccess uint32) (mslsad.LSAPR_HANDLE, error) {
	name := dtyp.NewUnicodeString(trustedDomainName)
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
