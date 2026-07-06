package functions

// IDL source: [MS-CMPO] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmpo/3a4677f0-9aef-41f9-9bca-f9c2469cefa6
// A fetched copy is kept at ms-cmpo.idl in the interface directory.

import (
	"fmt"

	IXnRemote "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/906b0ce0-c70b-1067-b317-00dd010662da/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmpo "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmpo"
)

// negotiateResourcesRequest carries the [in]/[in,out] parameters of NegotiateResources
// ([MS-CMPO] 3.4.4.3). phContext is a 20-byte context handle; pdwcAccepted is [in,out] and
// so also appears in the response.
type negotiateResourcesRequest struct {
	PhContext    mscmpo.PCONTEXT_HANDLE
	ResourceType mscmpo.RESOURCE_TYPE `ndr:"enum"`
	DwcRequested ndr.DWORD
	PdwcAccepted ndr.DWORD
}

func (*negotiateResourcesRequest) Opnum() uint16 { return IXnRemote.OpnumNegotiateResources }

// negotiateResourcesResponse carries the [in,out] accepted count and the HRESULT return
// value of NegotiateResources.
type negotiateResourcesResponse struct {
	PdwcAccepted ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// NegotiateResources calls NegotiateResources (opnum 2) ([MS-CMPO] 3.4.4.3): the caller
// requests a number of resources of resourceType on the session bound to phContext. It
// returns the number the partner accepted.
func NegotiateResources(rpc ndr.Invoker, phContext mscmpo.PCONTEXT_HANDLE, resourceType mscmpo.RESOURCE_TYPE, requested, accepted uint32) (uint32, error) {
	req := &negotiateResourcesRequest{
		PhContext:    phContext,
		ResourceType: resourceType,
		DwcRequested: ndr.DWORD(requested),
		PdwcAccepted: ndr.DWORD(accepted),
	}
	var resp negotiateResourcesResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("NegotiateResources: %w", err)
	}
	if uint32(resp.Status) != IXnRemote.StatusSuccess {
		return uint32(resp.PdwcAccepted), fmt.Errorf("NegotiateResources failed: %s", IXnRemote.StatusString(uint32(resp.Status)))
	}
	return uint32(resp.PdwcAccepted), nil
}
