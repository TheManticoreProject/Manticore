package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV6CreateClientInfoRequest carries the [in] parameters of R_DhcpV6CreateClientInfo.
type r_DhcpV6CreateClientInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_CLIENT_INFO_V6
}

func (*r_DhcpV6CreateClientInfoRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV6CreateClientInfo }

// r_DhcpV6CreateClientInfoResponse carries the [out] parameters and return value of R_DhcpV6CreateClientInfo.
type r_DhcpV6CreateClientInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV6CreateClientInfo calls R_DhcpV6CreateClientInfo (opnum 124) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV6CreateClientInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_CLIENT_INFO_V6) (err error) {
	req := &r_DhcpV6CreateClientInfoRequest{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpV6CreateClientInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV6CreateClientInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV6CreateClientInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
