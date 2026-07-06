package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrCreateUser2InDomainRequest carries the [in] parameters of SamrCreateUser2InDomain: the
// domain handle, the [ref] account name, the account type (a USER_ACCOUNT_* control bit), and
// the desired access mask for the returned user handle.
type samrCreateUser2InDomainRequest struct {
	DomainHandle  mssamr.SAMPR_HANDLE
	Name          msdtyp.RPC_UNICODE_STRING
	AccountType   ndr.DWORD
	DesiredAccess ndr.DWORD
}

func (*samrCreateUser2InDomainRequest) Opnum() uint16 { return samr.OpnumSamrCreateUser2InDomain }

// samrCreateUser2InDomainResponse is the reply: the [out] user handle, the access actually
// granted, the relative id assigned to the new user, and the NTSTATUS.
type samrCreateUser2InDomainResponse struct {
	UserHandle    mssamr.SAMPR_HANDLE
	GrantedAccess ndr.DWORD
	RelativeId    ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// SamrCreateUser2InDomain calls SamrCreateUser2InDomain (opnum 50), creating a user object of
// the given account type in the domain ([MS-SAMR] 3.1.5.4.4).
func SamrCreateUser2InDomain(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, name string, accountType uint32, desiredAccess uint32) (mssamr.SAMPR_HANDLE, uint32, uint32, error) {
	req := &samrCreateUser2InDomainRequest{
		DomainHandle:  domainHandle,
		Name:          msdtyp.NewUnicodeString(name),
		AccountType:   ndr.DWORD(accountType),
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp samrCreateUser2InDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.SAMPR_HANDLE{}, 0, 0, fmt.Errorf("SamrCreateUser2InDomain: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.UserHandle, uint32(resp.GrantedAccess), uint32(resp.RelativeId), fmt.Errorf("SamrCreateUser2InDomain failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.UserHandle, uint32(resp.GrantedAccess), uint32(resp.RelativeId), nil
}
