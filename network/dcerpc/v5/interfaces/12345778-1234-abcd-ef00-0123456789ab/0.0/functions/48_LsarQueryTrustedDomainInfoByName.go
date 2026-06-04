package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarQueryTrustedDomainInfoByNameRequest is the [in] parameter set of
// LsarQueryTrustedDomainInfoByName: an open policy handle, the [unique] trusted-domain name,
// and the information class selecting which union arm is returned.
type lsarQueryTrustedDomainInfoByNameRequest struct {
	PolicyHandle      structures.LSAPR_HANDLE
	TrustedDomainName dtyp.RPC_UNICODE_STRING
	InformationClass  structures.TRUSTED_INFORMATION_CLASS
}

func (*lsarQueryTrustedDomainInfoByNameRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarQueryTrustedDomainInfoByName
}

// lsarQueryTrustedDomainInfoByNameResponse is the reply: the [out, switch_is] union (a double
// pointer in the IDL, so a [unique] pointer on the wire) and the NTSTATUS return value.
type lsarQueryTrustedDomainInfoByNameResponse struct {
	TrustedDomainInformation *structures.LSAPR_TRUSTED_DOMAIN_INFO `ndr:"unique"`
	Status                   ndr.DWORD                             `ndr:"retval"`
}

// LsarQueryTrustedDomainInfoByName calls LsarQueryTrustedDomainInfoByName (opnum 48),
// returning the requested trusted-domain information class for the domain identified by name
// ([MS-LSAD] 3.1.4.7.6).
func LsarQueryTrustedDomainInfoByName(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, trustedDomainName string, infoClass structures.TRUSTED_INFORMATION_CLASS) (*structures.LSAPR_TRUSTED_DOMAIN_INFO, error) {
	name := dtyp.NewUnicodeString(trustedDomainName)
	req := &lsarQueryTrustedDomainInfoByNameRequest{
		PolicyHandle:      policyHandle,
		TrustedDomainName: name,
		InformationClass:  infoClass,
	}
	var resp lsarQueryTrustedDomainInfoByNameResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarQueryTrustedDomainInfoByName: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.TrustedDomainInformation, fmt.Errorf("LsarQueryTrustedDomainInfoByName failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.TrustedDomainInformation, nil
}
