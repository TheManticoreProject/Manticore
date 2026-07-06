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

// samrOpenDomainRequest is the [in] parameter set of SamrOpenDomain: a server handle, the
// desired access mask, and the [ref] SID of the domain to open (inline, single pointer).
type samrOpenDomainRequest struct {
	ServerHandle  mssamr.SAMPR_HANDLE
	DesiredAccess ndr.DWORD
	DomainId      msdtyp.RPC_SID
}

func (*samrOpenDomainRequest) Opnum() uint16 { return samr.OpnumSamrOpenDomain }

// SamrOpenDomain calls SamrOpenDomain (opnum 7), obtaining a handle to a domain object
// given its SID ([MS-SAMR] 3.1.5.1.5).
func SamrOpenDomain(rpc ndr.Invoker, serverHandle mssamr.SAMPR_HANDLE, desiredAccess uint32, domainId msdtyp.RPC_SID) (mssamr.SAMPR_HANDLE, error) {
	req := &samrOpenDomainRequest{
		ServerHandle:  serverHandle,
		DesiredAccess: ndr.DWORD(desiredAccess),
		DomainId:      domainId,
	}
	var resp openHandleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssamr.SAMPR_HANDLE{}, fmt.Errorf("SamrOpenDomain: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrOpenDomain failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
