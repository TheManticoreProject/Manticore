package functions

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrCreateAliasInDomainRequest carries the [in] parameters of SamrCreateAliasInDomain:
// the domain handle, the [ref] alias account name, and the desired access mask for the
// returned alias handle.
type samrCreateAliasInDomainRequest struct {
	DomainHandle  mssamr.SAMPR_HANDLE
	AccountName   msdtyp.RPC_UNICODE_STRING
	DesiredAccess ndr.DWORD
}

func (*samrCreateAliasInDomainRequest) Opnum() uint16 { return samr.OpnumSamrCreateAliasInDomain }

// samrCreateAliasInDomainResponse is the reply: the [out] alias handle, the new alias's
// relative id, and the NTSTATUS.
type samrCreateAliasInDomainResponse struct {
	AliasHandle mssamr.SAMPR_HANDLE
	RelativeId  ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// SamrCreateAliasInDomain calls SamrCreateAliasInDomain (opnum 14), creating an alias in
// the domain and returning a handle to it plus its RID ([MS-SAMR] 3.1.5.4.1).
func SamrCreateAliasInDomain(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, accountName string, desiredAccess uint32) (mssamr.SAMPR_HANDLE, uint32, error) {
	req := &samrCreateAliasInDomainRequest{
		DomainHandle:  domainHandle,
		AccountName:   msdtyp.NewUnicodeString(accountName),
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp samrCreateAliasInDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.SAMPR_HANDLE{}, 0, fmt.Errorf("SamrCreateAliasInDomain: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.AliasHandle, uint32(resp.RelativeId), fmt.Errorf("SamrCreateAliasInDomain failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.AliasHandle, uint32(resp.RelativeId), nil
}
