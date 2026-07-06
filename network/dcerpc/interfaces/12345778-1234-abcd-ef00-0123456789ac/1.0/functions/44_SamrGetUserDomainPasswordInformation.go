package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrGetUserDomainPasswordInformationRequest carries the [in] user handle whose domain
// password policy is queried.
type samrGetUserDomainPasswordInformationRequest struct {
	UserHandle mssamr.SAMPR_HANDLE
}

func (*samrGetUserDomainPasswordInformationRequest) Opnum() uint16 {
	return samr.OpnumSamrGetUserDomainPasswordInformation
}

// samrGetUserDomainPasswordInformationResponse is the reply: the [out] domain password policy
// (inline, single [ref] pointer) and the NTSTATUS.
type samrGetUserDomainPasswordInformationResponse struct {
	PasswordInformation mssamr.USER_DOMAIN_PASSWORD_INFORMATION
	Status              ndr.DWORD `ndr:"retval"`
}

// SamrGetUserDomainPasswordInformation calls SamrGetUserDomainPasswordInformation (opnum 44),
// returning the password policy of the domain the user belongs to ([MS-SAMR] 3.1.5.13.1).
func SamrGetUserDomainPasswordInformation(rpc ndr.Invoker, userHandle mssamr.SAMPR_HANDLE) (mssamr.USER_DOMAIN_PASSWORD_INFORMATION, error) {
	req := &samrGetUserDomainPasswordInformationRequest{UserHandle: userHandle}
	var resp samrGetUserDomainPasswordInformationResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.USER_DOMAIN_PASSWORD_INFORMATION{}, fmt.Errorf("SamrGetUserDomainPasswordInformation: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.PasswordInformation, fmt.Errorf("SamrGetUserDomainPasswordInformation failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.PasswordInformation, nil
}
