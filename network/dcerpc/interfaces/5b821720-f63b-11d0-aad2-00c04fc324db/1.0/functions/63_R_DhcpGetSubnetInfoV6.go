package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetSubnetInfoV6Request carries the [in] parameters of R_DhcpGetSubnetInfoV6.
type r_DhcpGetSubnetInfoV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   msdhcpm.DHCP_IPV6_ADDRESS
}

func (*r_DhcpGetSubnetInfoV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetSubnetInfoV6 }

// r_DhcpGetSubnetInfoV6Response carries the [out] parameters and return value of R_DhcpGetSubnetInfoV6.
type r_DhcpGetSubnetInfoV6Response struct {
	SubnetInfo *msdhcpm.DHCP_SUBNET_INFO_V6 `ndr:"unique"`
	Status     ndr.DWORD                    `ndr:"retval"`
}

// R_DhcpGetSubnetInfoV6 calls R_DhcpGetSubnetInfoV6 (opnum 63) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetSubnetInfoV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress msdhcpm.DHCP_IPV6_ADDRESS) (SubnetInfo *msdhcpm.DHCP_SUBNET_INFO_V6, err error) {
	req := &r_DhcpGetSubnetInfoV6Request{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
	}
	var resp r_DhcpGetSubnetInfoV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetSubnetInfoV6: %w", err)
		return
	}
	SubnetInfo = resp.SubnetInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetSubnetInfoV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
