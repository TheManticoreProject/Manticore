package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_EnumGlobalRoutingInfoRequest carries the [in] parameters of FAX_EnumGlobalRoutingInfo.
type fAX_EnumGlobalRoutingInfoRequest struct {
}

func (*fAX_EnumGlobalRoutingInfoRequest) Opnum() uint16 { return fax.OpnumFAX_EnumGlobalRoutingInfo }

// fAX_EnumGlobalRoutingInfoResponse carries the [out] parameters and return value of FAX_EnumGlobalRoutingInfo.
type fAX_EnumGlobalRoutingInfoResponse struct {
	RoutingInfoBuffer     []byte `ndr:"unique,conformant"`
	RoutingInfoBufferSize ndr.DWORD
	MethodsReturned       ndr.DWORD
	Status                ndr.DWORD `ndr:"retval"`
}

// FAX_EnumGlobalRoutingInfo calls FAX_EnumGlobalRoutingInfo (opnum 17) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EnumGlobalRoutingInfo(rpc ndr.Invoker) (RoutingInfoBuffer []byte, RoutingInfoBufferSize ndr.DWORD, MethodsReturned ndr.DWORD, err error) {
	req := &fAX_EnumGlobalRoutingInfoRequest{}
	var resp fAX_EnumGlobalRoutingInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EnumGlobalRoutingInfo: %w", err)
		return
	}
	RoutingInfoBuffer = resp.RoutingInfoBuffer
	RoutingInfoBufferSize = resp.RoutingInfoBufferSize
	MethodsReturned = resp.MethodsReturned
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EnumGlobalRoutingInfo failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
