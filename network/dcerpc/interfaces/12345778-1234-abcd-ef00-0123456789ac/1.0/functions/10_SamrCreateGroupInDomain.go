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

// samrCreateGroupInDomainRequest carries the [in] parameters of SamrCreateGroupInDomain:
// the domain handle, the [ref] group name, and the desired access mask for the returned
// group handle.
type samrCreateGroupInDomainRequest struct {
	DomainHandle  mssamr.SAMPR_HANDLE
	Name          msdtyp.RPC_UNICODE_STRING
	DesiredAccess ndr.DWORD
}

func (*samrCreateGroupInDomainRequest) Opnum() uint16 { return samr.OpnumSamrCreateGroupInDomain }

// samrCreateGroupInDomainResponse is the reply: the [out] group handle, the relative id
// assigned to the new group, and the NTSTATUS.
type samrCreateGroupInDomainResponse struct {
	GroupHandle mssamr.SAMPR_HANDLE
	RelativeId  ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// SamrCreateGroupInDomain calls SamrCreateGroupInDomain (opnum 10), creating a group object
// in the given domain ([MS-SAMR] 3.1.5.4.1).
func SamrCreateGroupInDomain(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, name string, desiredAccess uint32) (mssamr.SAMPR_HANDLE, uint32, error) {
	req := &samrCreateGroupInDomainRequest{
		DomainHandle:  domainHandle,
		Name:          msdtyp.NewUnicodeString(name),
		DesiredAccess: ndr.DWORD(desiredAccess),
	}
	var resp samrCreateGroupInDomainResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.SAMPR_HANDLE{}, 0, fmt.Errorf("SamrCreateGroupInDomain: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.GroupHandle, uint32(resp.RelativeId), fmt.Errorf("SamrCreateGroupInDomain failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.GroupHandle, uint32(resp.RelativeId), nil
}
