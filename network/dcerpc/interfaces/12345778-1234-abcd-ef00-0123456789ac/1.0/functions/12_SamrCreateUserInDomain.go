package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrCreateUserInDomainRequest carries the [in] parameters of SamrCreateUserInDomain: the
// domain handle, the [ref] account name, and the desired access mask for the returned user
// handle.
type samrCreateUserInDomainRequest struct {
	DomainHandle  mssamr.SAMPR_HANDLE
	Name          dtyp.RPC_UNICODE_STRING
	DesiredAccess ndr.DWORD
}

func (*samrCreateUserInDomainRequest) Opnum() uint16 { return samr.OpnumSamrCreateUserInDomain }

// samrCreateUserInDomainResponse is the reply: the [out] user handle, the relative id
// assigned to the new user, and the NTSTATUS.
type samrCreateUserInDomainResponse struct {
	UserHandle mssamr.SAMPR_HANDLE
	RelativeId ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// SamrCreateUserInDomain calls SamrCreateUserInDomain (opnum 12), creating a user object in
// the given domain ([MS-SAMR] 3.1.5.4.4).
func SamrCreateUserInDomain(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, name string, desiredAccess uint32) (mssamr.SAMPR_HANDLE, uint32, error) {
	req := &samrCreateUserInDomainRequest{
		DomainHandle:  domainHandle,
		Name:          dtyp.NewUnicodeString(name),
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp samrCreateUserInDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.SAMPR_HANDLE{}, 0, fmt.Errorf("SamrCreateUserInDomain: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.UserHandle, uint32(resp.RelativeId), fmt.Errorf("SamrCreateUserInDomain failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.UserHandle, uint32(resp.RelativeId), nil
}
