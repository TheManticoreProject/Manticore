package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// lsarQueryTrustedDomainInfoRequest is the [in] parameter set of LsarQueryTrustedDomainInfo:
// an open policy handle, the [unique] SID identifying the trusted domain, and the information
// class selecting which union arm is returned.
type lsarQueryTrustedDomainInfoRequest struct {
	PolicyHandle     structures.LSAPR_HANDLE
	TrustedDomainSid *dtyp.RPC_SID                        `ndr:"unique"`
	InformationClass structures.TRUSTED_INFORMATION_CLASS `ndr:"enum"`
}

func (*lsarQueryTrustedDomainInfoRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarQueryTrustedDomainInfo
}

// lsarQueryTrustedDomainInfoResponse is the reply: the [out, switch_is] union (a double
// pointer in the IDL, so a [unique] pointer on the wire) and the NTSTATUS return value.
type lsarQueryTrustedDomainInfoResponse struct {
	TrustedDomainInformation *structures.LSAPR_TRUSTED_DOMAIN_INFO `ndr:"unique"`
	Status                   ndr.DWORD                             `ndr:"retval"`
}

// LsarQueryTrustedDomainInfo calls LsarQueryTrustedDomainInfo (opnum 39), returning the
// requested trusted-domain information class for the domain identified by SID
// ([MS-LSAD] 3.1.4.7.2).
func LsarQueryTrustedDomainInfo(rpc ndr.Invoker, policyHandle structures.LSAPR_HANDLE, trustedDomainSid *dtyp.RPC_SID, infoClass structures.TRUSTED_INFORMATION_CLASS) (*structures.LSAPR_TRUSTED_DOMAIN_INFO, error) {
	req := &lsarQueryTrustedDomainInfoRequest{
		PolicyHandle:     policyHandle,
		TrustedDomainSid: trustedDomainSid,
		InformationClass: infoClass,
	}
	var resp lsarQueryTrustedDomainInfoResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarQueryTrustedDomainInfo: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.TrustedDomainInformation, fmt.Errorf("LsarQueryTrustedDomainInfo failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.TrustedDomainInformation, nil
}
