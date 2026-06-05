package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrLookupDomainInSamServerRequest carries the [in] parameters of
// SamrLookupDomainInSamServer: the server handle and the [ref] domain name to
// resolve (modeled inline as an RPC_UNICODE_STRING).
type samrLookupDomainInSamServerRequest struct {
	ServerHandle structures.SAMPR_HANDLE
	Name         dtyp.RPC_UNICODE_STRING
}

func (*samrLookupDomainInSamServerRequest) Opnum() uint16 {
	return samr.OpnumSamrLookupDomainInSamServer
}

// samrLookupDomainInSamServerResponse carries the [out] double pointer to the
// resolved domain SID and the NTSTATUS.
type samrLookupDomainInSamServerResponse struct {
	DomainId *dtyp.RPC_SID `ndr:"unique"`
	Status   ndr.DWORD     `ndr:"retval"`
}

// SamrLookupDomainInSamServer calls SamrLookupDomainInSamServer (opnum 5),
// resolving a domain name to its SID ([MS-SAMR] 3.1.5.11.1).
func SamrLookupDomainInSamServer(rpc ndr.Invoker, handle structures.SAMPR_HANDLE, name string) (*dtyp.RPC_SID, error) {
	req := &samrLookupDomainInSamServerRequest{
		ServerHandle: handle,
		Name:         dtyp.NewUnicodeString(name),
	}
	var resp samrLookupDomainInSamServerResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrLookupDomainInSamServer: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.DomainId, fmt.Errorf("SamrLookupDomainInSamServer failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.DomainId, nil
}
