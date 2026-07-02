package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_EnumRoutingMethodsRequest carries the [in] parameters of FAX_EnumRoutingMethods.
type fAX_EnumRoutingMethodsRequest struct {
	FaxPortHandle msfax.RPC_FAX_PORT_HANDLE
}

func (*fAX_EnumRoutingMethodsRequest) Opnum() uint16 { return fax.OpnumFAX_EnumRoutingMethods }

// fAX_EnumRoutingMethodsResponse carries the [out] parameters and return value of FAX_EnumRoutingMethods.
type fAX_EnumRoutingMethodsResponse struct {
	RoutingInfoBuffer     []byte `ndr:"unique,conformant"`
	RoutingInfoBufferSize ndr.DWORD
	PortsReturned         ndr.DWORD
	Status                ndr.DWORD `ndr:"retval"`
}

// FAX_EnumRoutingMethods calls FAX_EnumRoutingMethods (opnum 13) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EnumRoutingMethods(rpc ndr.Invoker, faxPortHandle msfax.RPC_FAX_PORT_HANDLE) (RoutingInfoBuffer []byte, RoutingInfoBufferSize ndr.DWORD, PortsReturned ndr.DWORD, err error) {
	req := &fAX_EnumRoutingMethodsRequest{
		FaxPortHandle: faxPortHandle,
	}
	var resp fAX_EnumRoutingMethodsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EnumRoutingMethods: %w", err)
		return
	}
	RoutingInfoBuffer = resp.RoutingInfoBuffer
	RoutingInfoBufferSize = resp.RoutingInfoBufferSize
	PortsReturned = resp.PortsReturned
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EnumRoutingMethods failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
