package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetSubnetInfoRequest carries the [in] parameters of R_DhcpGetSubnetInfo.
type r_DhcpGetSubnetInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
}

func (*r_DhcpGetSubnetInfoRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpGetSubnetInfo }

// r_DhcpGetSubnetInfoResponse carries the [out] parameters and return value of R_DhcpGetSubnetInfo.
type r_DhcpGetSubnetInfoResponse struct {
	SubnetInfo *msdhcpm.DHCP_SUBNET_INFO `ndr:"unique"`
	Status     ndr.DWORD                 `ndr:"retval"`
}

// R_DhcpGetSubnetInfo calls R_DhcpGetSubnetInfo (opnum 2) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetSubnetInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD) (SubnetInfo *msdhcpm.DHCP_SUBNET_INFO, err error) {
	req := &r_DhcpGetSubnetInfoRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
	}
	var resp r_DhcpGetSubnetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetSubnetInfo: %w", err)
		return
	}
	SubnetInfo = resp.SubnetInfo
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetSubnetInfo failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
