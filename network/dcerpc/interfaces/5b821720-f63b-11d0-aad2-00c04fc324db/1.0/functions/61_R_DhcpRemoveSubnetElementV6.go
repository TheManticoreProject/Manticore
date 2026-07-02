package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpRemoveSubnetElementV6Request carries the [in] parameters of R_DhcpRemoveSubnetElementV6.
type r_DhcpRemoveSubnetElementV6Request struct {
	ServerIpAddress   *ndr.WSTR `ndr:"unique"`
	SubnetAddress     msdhcpm.DHCP_IPV6_ADDRESS
	RemoveElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V6
	ForceFlag         msdhcpm.DHCP_FORCE_FLAG
}

func (*r_DhcpRemoveSubnetElementV6Request) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpRemoveSubnetElementV6
}

// r_DhcpRemoveSubnetElementV6Response carries the [out] parameters and return value of R_DhcpRemoveSubnetElementV6.
type r_DhcpRemoveSubnetElementV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpRemoveSubnetElementV6 calls R_DhcpRemoveSubnetElementV6 (opnum 61) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpRemoveSubnetElementV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress msdhcpm.DHCP_IPV6_ADDRESS, removeElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V6, forceFlag msdhcpm.DHCP_FORCE_FLAG) (err error) {
	req := &r_DhcpRemoveSubnetElementV6Request{
		ServerIpAddress:   serverIpAddress,
		SubnetAddress:     subnetAddress,
		RemoveElementInfo: removeElementInfo,
		ForceFlag:         forceFlag,
	}
	var resp r_DhcpRemoveSubnetElementV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpRemoveSubnetElementV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpRemoveSubnetElementV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
