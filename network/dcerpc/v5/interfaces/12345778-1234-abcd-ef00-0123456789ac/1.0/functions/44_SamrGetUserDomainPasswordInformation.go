package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrGetUserDomainPasswordInformationRequest carries the [in] user handle whose domain
// password policy is queried.
type samrGetUserDomainPasswordInformationRequest struct {
	UserHandle structures.SAMPR_HANDLE
}

func (*samrGetUserDomainPasswordInformationRequest) Opnum() uint16 {
	return samr.OpnumSamrGetUserDomainPasswordInformation
}

// samrGetUserDomainPasswordInformationResponse is the reply: the [out] domain password policy
// (inline, single [ref] pointer) and the NTSTATUS.
type samrGetUserDomainPasswordInformationResponse struct {
	PasswordInformation structures.USER_DOMAIN_PASSWORD_INFORMATION
	Status              ndr.DWORD `ndr:"retval"`
}

// SamrGetUserDomainPasswordInformation calls SamrGetUserDomainPasswordInformation (opnum 44),
// returning the password policy of the domain the user belongs to ([MS-SAMR] 3.1.5.13.1).
func SamrGetUserDomainPasswordInformation(rpc *client.Client, userHandle structures.SAMPR_HANDLE) (structures.USER_DOMAIN_PASSWORD_INFORMATION, error) {
	req := &samrGetUserDomainPasswordInformationRequest{UserHandle: userHandle}
	var resp samrGetUserDomainPasswordInformationResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.USER_DOMAIN_PASSWORD_INFORMATION{}, fmt.Errorf("SamrGetUserDomainPasswordInformation: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.PasswordInformation, fmt.Errorf("SamrGetUserDomainPasswordInformation failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.PasswordInformation, nil
}
