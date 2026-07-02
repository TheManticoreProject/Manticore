package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4CreateClientInfoExRequest carries the [in] parameters of R_DhcpV4CreateClientInfoEx.
type r_DhcpV4CreateClientInfoExRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_CLIENT_INFO_EX
}

func (*r_DhcpV4CreateClientInfoExRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4CreateClientInfoEx
}

// r_DhcpV4CreateClientInfoExResponse carries the [out] parameters and return value of R_DhcpV4CreateClientInfoEx.
type r_DhcpV4CreateClientInfoExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4CreateClientInfoEx calls R_DhcpV4CreateClientInfoEx (opnum 131) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4CreateClientInfoEx(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_CLIENT_INFO_EX) (err error) {
	req := &r_DhcpV4CreateClientInfoExRequest{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpV4CreateClientInfoExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4CreateClientInfoEx: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4CreateClientInfoEx failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
