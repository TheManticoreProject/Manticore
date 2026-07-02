package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4CreateClientInfoRequest carries the [in] parameters of R_DhcpV4CreateClientInfo.
type r_DhcpV4CreateClientInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_CLIENT_INFO_PB
}

func (*r_DhcpV4CreateClientInfoRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4CreateClientInfo }

// r_DhcpV4CreateClientInfoResponse carries the [out] parameters and return value of R_DhcpV4CreateClientInfo.
type r_DhcpV4CreateClientInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4CreateClientInfo calls R_DhcpV4CreateClientInfo (opnum 122) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4CreateClientInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_CLIENT_INFO_PB) (err error) {
	req := &r_DhcpV4CreateClientInfoRequest{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpV4CreateClientInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4CreateClientInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4CreateClientInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
