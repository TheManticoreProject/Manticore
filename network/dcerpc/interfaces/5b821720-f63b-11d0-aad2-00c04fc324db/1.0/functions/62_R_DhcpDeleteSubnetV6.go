package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpDeleteSubnetV6Request carries the [in] parameters of R_DhcpDeleteSubnetV6.
type r_DhcpDeleteSubnetV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   msdhcpm.DHCP_IPV6_ADDRESS
	ForceFlag       msdhcpm.DHCP_FORCE_FLAG
}

func (*r_DhcpDeleteSubnetV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpDeleteSubnetV6 }

// r_DhcpDeleteSubnetV6Response carries the [out] parameters and return value of R_DhcpDeleteSubnetV6.
type r_DhcpDeleteSubnetV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpDeleteSubnetV6 calls R_DhcpDeleteSubnetV6 (opnum 62) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpDeleteSubnetV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress msdhcpm.DHCP_IPV6_ADDRESS, forceFlag msdhcpm.DHCP_FORCE_FLAG) (err error) {
	req := &r_DhcpDeleteSubnetV6Request{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		ForceFlag:       forceFlag,
	}
	var resp r_DhcpDeleteSubnetV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpDeleteSubnetV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpDeleteSubnetV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
