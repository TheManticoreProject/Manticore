package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpAddSubnetElementV6Request carries the [in] parameters of R_DhcpAddSubnetElementV6.
type r_DhcpAddSubnetElementV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   msdhcpm.DHCP_IPV6_ADDRESS
	AddElementInfo  msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V6
}

func (*r_DhcpAddSubnetElementV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpAddSubnetElementV6 }

// r_DhcpAddSubnetElementV6Response carries the [out] parameters and return value of R_DhcpAddSubnetElementV6.
type r_DhcpAddSubnetElementV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpAddSubnetElementV6 calls R_DhcpAddSubnetElementV6 (opnum 59) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpAddSubnetElementV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress msdhcpm.DHCP_IPV6_ADDRESS, addElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V6) (err error) {
	req := &r_DhcpAddSubnetElementV6Request{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		AddElementInfo:  addElementInfo,
	}
	var resp r_DhcpAddSubnetElementV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpAddSubnetElementV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpAddSubnetElementV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
