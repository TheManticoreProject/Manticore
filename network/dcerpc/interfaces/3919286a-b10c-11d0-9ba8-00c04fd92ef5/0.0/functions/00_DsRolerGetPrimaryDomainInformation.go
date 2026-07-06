package functions

// IDL source: [MS-DSSP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dssp/ae6edaa5-b40e-4cc9-9ebc-42cc657ce61e
// A fetched copy is kept at ms-dssp.idl in the interface directory.

import (
	"fmt"

	dssetup "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/3919286a-b10c-11d0-9ba8-00c04fd92ef5/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdssp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dssp"
)

// dsRolerGetPrimaryDomainInformationRequest carries the [in] parameters of DsRolerGetPrimaryDomainInformation.
type dsRolerGetPrimaryDomainInformationRequest struct {
	InfoLevel msdssp.DSROLE_PRIMARY_DOMAIN_INFO_LEVEL
}

func (*dsRolerGetPrimaryDomainInformationRequest) Opnum() uint16 {
	return dssetup.OpnumDsRolerGetPrimaryDomainInformation
}

// dsRolerGetPrimaryDomainInformationResponse carries the [out] parameter and return
// value of DsRolerGetPrimaryDomainInformation.
//
// The IDL declares DomainInfo as a double pointer, [out, switch_is(InfoLevel)]
// PDSROLER_PRIMARY_DOMAIN_INFORMATION *DomainInfo. The outer pointer is the top-level
// [out] parameter reference (transmitted in place, no referent id); the inner PDSROLER_*
// pointer takes the interface's unique default and emits a referent id, so the wire shape
// is a single [unique] pointer to the union — modeled here as *T with ndr:"unique". The
// union's discriminant (an NDR enum) precedes the selected arm on the wire.
type dsRolerGetPrimaryDomainInformationResponse struct {
	DomainInfo *msdssp.DSROLER_PRIMARY_DOMAIN_INFORMATION `ndr:"unique"`
	Status     ndr.DWORD                                  `ndr:"retval"`
}

// DsRolerGetPrimaryDomainInformation calls DsRolerGetPrimaryDomainInformation (opnum 0)
// ([MS-DSSP] 3.2.5.1). It returns the requested level of domain-membership information
// for the server. The caller selects the level via infoLevel; the returned union arm is
// discriminated by the same value (inspect DomainInfo.Tag).
func DsRolerGetPrimaryDomainInformation(rpc ndr.Invoker, infoLevel msdssp.DSROLE_PRIMARY_DOMAIN_INFO_LEVEL) (DomainInfo *msdssp.DSROLER_PRIMARY_DOMAIN_INFORMATION, err error) {
	req := &dsRolerGetPrimaryDomainInformationRequest{
		InfoLevel: infoLevel,
	}
	var resp dsRolerGetPrimaryDomainInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsRolerGetPrimaryDomainInformation: %w", err)
		return
	}
	DomainInfo = resp.DomainInfo
	if uint32(resp.Status) != dssetup.StatusSuccess {
		err = fmt.Errorf("DsRolerGetPrimaryDomainInformation failed: %s", dssetup.StatusString(uint32(resp.Status)))
	}
	return
}
