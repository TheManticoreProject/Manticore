package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrGetDomainPasswordInformationRequest carries the [in] parameters. The
// handle_t binding handle is implicit (the RPC client) and is not marshalled;
// Unused is an ignored [unique] string per the protocol.
type samrGetDomainPasswordInformationRequest struct {
	Unused *dtyp.RPC_UNICODE_STRING `ndr:"unique"`
}

func (*samrGetDomainPasswordInformationRequest) Opnum() uint16 {
	return samr.OpnumSamrGetDomainPasswordInformation
}

// samrGetDomainPasswordInformationResponse carries the [out, ref] password policy
// information (modeled inline) and the NTSTATUS.
type samrGetDomainPasswordInformationResponse struct {
	PasswordInformation structures.USER_DOMAIN_PASSWORD_INFORMATION
	Status              ndr.DWORD `ndr:"retval"`
}

// SamrGetDomainPasswordInformation calls SamrGetDomainPasswordInformation
// (opnum 56), retrieving select password policy properties of the account domain
// ([MS-SAMR] 3.1.5.13.3).
func SamrGetDomainPasswordInformation(rpc *client.Client) (structures.USER_DOMAIN_PASSWORD_INFORMATION, error) {
	req := &samrGetDomainPasswordInformationRequest{}
	var resp samrGetDomainPasswordInformationResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.USER_DOMAIN_PASSWORD_INFORMATION{}, fmt.Errorf("SamrGetDomainPasswordInformation: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.PasswordInformation, fmt.Errorf("SamrGetDomainPasswordInformation failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.PasswordInformation, nil
}
