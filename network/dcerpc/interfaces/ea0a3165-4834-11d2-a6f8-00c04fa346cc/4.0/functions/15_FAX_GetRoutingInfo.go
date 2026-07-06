package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_GetRoutingInfoRequest carries the [in] parameters of FAX_GetRoutingInfo.
type fAX_GetRoutingInfoRequest struct {
	FaxPortHandle msfax.RPC_FAX_PORT_HANDLE
	RoutingGuid   *ndr.WSTR `ndr:"unique"`
}

func (*fAX_GetRoutingInfoRequest) Opnum() uint16 { return fax.OpnumFAX_GetRoutingInfo }

// fAX_GetRoutingInfoResponse carries the [out] parameters and return value of FAX_GetRoutingInfo.
type fAX_GetRoutingInfoResponse struct {
	RoutingInfoBuffer     []byte `ndr:"unique,conformant"`
	RoutingInfoBufferSize ndr.DWORD
	Status                ndr.DWORD `ndr:"retval"`
}

// FAX_GetRoutingInfo calls FAX_GetRoutingInfo (opnum 15) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetRoutingInfo(rpc ndr.Invoker, faxPortHandle msfax.RPC_FAX_PORT_HANDLE, routingGuid *ndr.WSTR) (RoutingInfoBuffer []byte, RoutingInfoBufferSize ndr.DWORD, err error) {
	req := &fAX_GetRoutingInfoRequest{
		FaxPortHandle: faxPortHandle,
		RoutingGuid:   routingGuid,
	}
	var resp fAX_GetRoutingInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetRoutingInfo: %w", err)
		return
	}
	RoutingInfoBuffer = resp.RoutingInfoBuffer
	RoutingInfoBufferSize = resp.RoutingInfoBufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetRoutingInfo failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
