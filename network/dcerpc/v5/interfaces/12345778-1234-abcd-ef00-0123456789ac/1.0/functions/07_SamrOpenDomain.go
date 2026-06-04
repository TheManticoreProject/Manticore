package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrOpenDomainRequest is the [in] parameter set of SamrOpenDomain: a server handle, the
// desired access mask, and the [ref] SID of the domain to open (inline, single pointer).
type samrOpenDomainRequest struct {
	ServerHandle  structures.SAMPR_HANDLE
	DesiredAccess ndr.DWORD
	DomainId      dtyp.RPC_SID
}

func (*samrOpenDomainRequest) Opnum() uint16 { return samr.OpnumSamrOpenDomain }

// SamrOpenDomain calls SamrOpenDomain (opnum 7), obtaining a handle to a domain object
// given its SID ([MS-SAMR] 3.1.5.1.5).
func SamrOpenDomain(rpc *client.Client, serverHandle structures.SAMPR_HANDLE, desiredAccess uint32, domainId dtyp.RPC_SID) (structures.SAMPR_HANDLE, error) {
	req := &samrOpenDomainRequest{
		ServerHandle:  serverHandle,
		DesiredAccess: ndr.DWORD(desiredAccess),
		DomainId:      domainId,
	}
	var resp openHandleResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SAMPR_HANDLE{}, fmt.Errorf("SamrOpenDomain: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.Handle, fmt.Errorf("SamrOpenDomain failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
