package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpServerRedoAuthorizationRequest carries the [in] parameters of R_DhcpServerRedoAuthorization.
type r_DhcpServerRedoAuthorizationRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	DwReserved      ndr.DWORD
}

func (*r_DhcpServerRedoAuthorizationRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpServerRedoAuthorization
}

// r_DhcpServerRedoAuthorizationResponse carries the [out] parameters and return value of R_DhcpServerRedoAuthorization.
type r_DhcpServerRedoAuthorizationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpServerRedoAuthorization calls R_DhcpServerRedoAuthorization (opnum 36) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerRedoAuthorization(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, dwReserved ndr.DWORD) (err error) {
	req := &r_DhcpServerRedoAuthorizationRequest{
		ServerIpAddress: serverIpAddress,
		DwReserved:      dwReserved,
	}
	var resp r_DhcpServerRedoAuthorizationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerRedoAuthorization: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerRedoAuthorization failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
