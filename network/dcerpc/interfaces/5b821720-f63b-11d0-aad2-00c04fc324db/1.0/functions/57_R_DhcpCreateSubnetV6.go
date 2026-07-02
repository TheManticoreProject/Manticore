package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpCreateSubnetV6Request carries the [in] parameters of R_DhcpCreateSubnetV6.
type r_DhcpCreateSubnetV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   msdhcpm.DHCP_IPV6_ADDRESS
	SubnetInfo      msdhcpm.DHCP_SUBNET_INFO_V6
}

func (*r_DhcpCreateSubnetV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpCreateSubnetV6 }

// r_DhcpCreateSubnetV6Response carries the [out] parameters and return value of R_DhcpCreateSubnetV6.
type r_DhcpCreateSubnetV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpCreateSubnetV6 calls R_DhcpCreateSubnetV6 (opnum 57) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpCreateSubnetV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress msdhcpm.DHCP_IPV6_ADDRESS, subnetInfo msdhcpm.DHCP_SUBNET_INFO_V6) (err error) {
	req := &r_DhcpCreateSubnetV6Request{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		SubnetInfo:      subnetInfo,
	}
	var resp r_DhcpCreateSubnetV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpCreateSubnetV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpCreateSubnetV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
