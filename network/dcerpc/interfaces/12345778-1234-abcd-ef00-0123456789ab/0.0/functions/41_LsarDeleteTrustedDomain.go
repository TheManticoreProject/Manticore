package functions

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarDeleteTrustedDomainRequest is the [in] parameter set of LsarDeleteTrustedDomain: an
// open policy handle and the [unique] SID of the trusted domain to delete.
type lsarDeleteTrustedDomainRequest struct {
	PolicyHandle     mslsad.LSAPR_HANDLE
	TrustedDomainSid *msdtyp.RPC_SID `ndr:"unique"`
}

func (*lsarDeleteTrustedDomainRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarDeleteTrustedDomain
}

// LsarDeleteTrustedDomain calls LsarDeleteTrustedDomain (opnum 41), deleting the trusted
// domain object identified by its SID.
func LsarDeleteTrustedDomain(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, trustedDomainSid *msdtyp.RPC_SID) error {
	req := &lsarDeleteTrustedDomainRequest{
		PolicyHandle:     policyHandle,
		TrustedDomainSid: trustedDomainSid,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarDeleteTrustedDomain: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarDeleteTrustedDomain failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
