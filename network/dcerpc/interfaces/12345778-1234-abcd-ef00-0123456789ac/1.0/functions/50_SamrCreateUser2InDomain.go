package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrCreateUser2InDomainRequest carries the [in] parameters of SamrCreateUser2InDomain: the
// domain handle, the [ref] account name, the account type (a USER_ACCOUNT_* control bit), and
// the desired access mask for the returned user handle.
type samrCreateUser2InDomainRequest struct {
	DomainHandle  structures.SAMPR_HANDLE
	Name          dtyp.RPC_UNICODE_STRING
	AccountType   ndr.DWORD
	DesiredAccess ndr.DWORD
}

func (*samrCreateUser2InDomainRequest) Opnum() uint16 { return samr.OpnumSamrCreateUser2InDomain }

// samrCreateUser2InDomainResponse is the reply: the [out] user handle, the access actually
// granted, the relative id assigned to the new user, and the NTSTATUS.
type samrCreateUser2InDomainResponse struct {
	UserHandle    structures.SAMPR_HANDLE
	GrantedAccess ndr.DWORD
	RelativeId    ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// SamrCreateUser2InDomain calls SamrCreateUser2InDomain (opnum 50), creating a user object of
// the given account type in the domain ([MS-SAMR] 3.1.5.4.4).
func SamrCreateUser2InDomain(rpc ndr.Invoker, domainHandle structures.SAMPR_HANDLE, name string, accountType uint32, desiredAccess uint32) (structures.SAMPR_HANDLE, uint32, uint32, error) {
	req := &samrCreateUser2InDomainRequest{
		DomainHandle:  domainHandle,
		Name:          dtyp.NewUnicodeString(name),
		AccountType:   ndr.DWORD(accountType),
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp samrCreateUser2InDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_HANDLE{}, 0, 0, fmt.Errorf("SamrCreateUser2InDomain: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.UserHandle, uint32(resp.GrantedAccess), uint32(resp.RelativeId), fmt.Errorf("SamrCreateUser2InDomain failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.UserHandle, uint32(resp.GrantedAccess), uint32(resp.RelativeId), nil
}
