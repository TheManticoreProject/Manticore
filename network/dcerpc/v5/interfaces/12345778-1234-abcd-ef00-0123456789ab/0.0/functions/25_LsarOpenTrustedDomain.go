package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarOpenTrustedDomainRequest is the [in] parameter set of LsarOpenTrustedDomain: an open
// policy handle, the [unique] SID of the trusted domain to open, and the desired access
// mask for the returned handle.
type lsarOpenTrustedDomainRequest struct {
	PolicyHandle     structures.LSAPR_HANDLE
	TrustedDomainSid *dtyp.RPC_SID `ndr:"unique"`
	DesiredAccess    ndr.DWORD
}

func (*lsarOpenTrustedDomainRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarOpenTrustedDomain
}

// LsarOpenTrustedDomain calls LsarOpenTrustedDomain (opnum 25), opening the trusted domain
// object identified by its SID and returning a handle to it.
func LsarOpenTrustedDomain(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, trustedDomainSid *dtyp.RPC_SID, desiredAccess uint32) (structures.LSAPR_HANDLE, error) {
	req := &lsarOpenTrustedDomainRequest{
		PolicyHandle:     policyHandle,
		TrustedDomainSid: trustedDomainSid,
		DesiredAccess:    ndr.DWORD(desiredAccess),
	}
	var resp handleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.LSAPR_HANDLE{}, fmt.Errorf("LsarOpenTrustedDomain: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.Handle, fmt.Errorf("LsarOpenTrustedDomain failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
