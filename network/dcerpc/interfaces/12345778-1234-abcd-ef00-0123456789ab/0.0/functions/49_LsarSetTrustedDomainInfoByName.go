package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarSetTrustedDomainInfoByNameRequest is the [in] parameter set of
// LsarSetTrustedDomainInfoByName: an open policy handle, the [unique] trusted-domain name,
// the information class, and the inline [ref] union value carrying the new trusted-domain
// information.
type lsarSetTrustedDomainInfoByNameRequest struct {
	PolicyHandle             mslsad.LSAPR_HANDLE
	TrustedDomainName        msdtyp.RPC_UNICODE_STRING
	InformationClass         mslsad.TRUSTED_INFORMATION_CLASS `ndr:"enum"`
	TrustedDomainInformation mslsad.LSAPR_TRUSTED_DOMAIN_INFO
}

func (*lsarSetTrustedDomainInfoByNameRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarSetTrustedDomainInfoByName
}

// LsarSetTrustedDomainInfoByName calls LsarSetTrustedDomainInfoByName (opnum 49), setting the
// trusted-domain information for the domain identified by name ([MS-LSAD] 3.1.4.7.7). The
// union discriminant is set to the information class before marshalling.
func LsarSetTrustedDomainInfoByName(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, trustedDomainName string, infoClass mslsad.TRUSTED_INFORMATION_CLASS, info mslsad.LSAPR_TRUSTED_DOMAIN_INFO) error {
	info.Class = infoClass
	name := msdtyp.NewUnicodeString(trustedDomainName)
	req := &lsarSetTrustedDomainInfoByNameRequest{
		PolicyHandle:             policyHandle,
		TrustedDomainName:        name,
		InformationClass:         infoClass,
		TrustedDomainInformation: info,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("LsarSetTrustedDomainInfoByName: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return fmt.Errorf("LsarSetTrustedDomainInfoByName failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return nil
}
