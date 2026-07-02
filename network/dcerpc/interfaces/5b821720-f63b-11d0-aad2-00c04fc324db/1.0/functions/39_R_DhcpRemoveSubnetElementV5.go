package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpRemoveSubnetElementV5Request carries the [in] parameters of R_DhcpRemoveSubnetElementV5.
type r_DhcpRemoveSubnetElementV5Request struct {
	ServerIpAddress   *ndr.WSTR `ndr:"unique"`
	SubnetAddress     ndr.DWORD
	RemoveElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V5
	ForceFlag         msdhcpm.DHCP_FORCE_FLAG
}

func (*r_DhcpRemoveSubnetElementV5Request) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpRemoveSubnetElementV5
}

// r_DhcpRemoveSubnetElementV5Response carries the [out] parameters and return value of R_DhcpRemoveSubnetElementV5.
type r_DhcpRemoveSubnetElementV5Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpRemoveSubnetElementV5 calls R_DhcpRemoveSubnetElementV5 (opnum 39) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpRemoveSubnetElementV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, removeElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V5, forceFlag msdhcpm.DHCP_FORCE_FLAG) (err error) {
	req := &r_DhcpRemoveSubnetElementV5Request{
		ServerIpAddress:   serverIpAddress,
		SubnetAddress:     subnetAddress,
		RemoveElementInfo: removeElementInfo,
		ForceFlag:         forceFlag,
	}
	var resp r_DhcpRemoveSubnetElementV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpRemoveSubnetElementV5: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpRemoveSubnetElementV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
