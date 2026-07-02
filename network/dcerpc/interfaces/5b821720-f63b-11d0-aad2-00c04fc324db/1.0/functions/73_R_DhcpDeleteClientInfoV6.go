package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpDeleteClientInfoV6Request carries the [in] parameters of R_DhcpDeleteClientInfoV6.
type r_DhcpDeleteClientInfoV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_SEARCH_INFO_V6
}

func (*r_DhcpDeleteClientInfoV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpDeleteClientInfoV6 }

// r_DhcpDeleteClientInfoV6Response carries the [out] parameters and return value of R_DhcpDeleteClientInfoV6.
type r_DhcpDeleteClientInfoV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpDeleteClientInfoV6 calls R_DhcpDeleteClientInfoV6 (opnum 73) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpDeleteClientInfoV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_SEARCH_INFO_V6) (err error) {
	req := &r_DhcpDeleteClientInfoV6Request{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpDeleteClientInfoV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpDeleteClientInfoV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpDeleteClientInfoV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
