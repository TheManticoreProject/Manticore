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

// samrLookupDomainInSamServerRequest carries the [in] parameters of
// SamrLookupDomainInSamServer: the server handle and the [ref] domain name to
// resolve (modeled inline as an RPC_UNICODE_STRING).
type samrLookupDomainInSamServerRequest struct {
	ServerHandle mssamr.SAMPR_HANDLE
	Name         msdtyp.RPC_UNICODE_STRING
}

func (*samrLookupDomainInSamServerRequest) Opnum() uint16 {
	return samr.OpnumSamrLookupDomainInSamServer
}

// samrLookupDomainInSamServerResponse carries the [out] double pointer to the
// resolved domain SID and the NTSTATUS.
type samrLookupDomainInSamServerResponse struct {
	DomainId *msdtyp.RPC_SID `ndr:"unique"`
	Status   ndr.DWORD       `ndr:"retval"`
}

// SamrLookupDomainInSamServer calls SamrLookupDomainInSamServer (opnum 5),
// resolving a domain name to its SID ([MS-SAMR] 3.1.5.11.1).
func SamrLookupDomainInSamServer(rpc ndr.Invoker, handle mssamr.SAMPR_HANDLE, name string) (*msdtyp.RPC_SID, error) {
	req := &samrLookupDomainInSamServerRequest{
		ServerHandle: handle,
		Name:         msdtyp.NewUnicodeString(name),
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
