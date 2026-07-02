package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetSubnetInfoVQRequest carries the [in] parameters of R_DhcpGetSubnetInfoVQ.
type r_DhcpGetSubnetInfoVQRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
}

func (*r_DhcpGetSubnetInfoVQRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpGetSubnetInfoVQ }

// r_DhcpGetSubnetInfoVQResponse carries the [out] parameters and return value of R_DhcpGetSubnetInfoVQ.
type r_DhcpGetSubnetInfoVQResponse struct {
	SubnetInfoVQ *msdhcpm.DHCP_SUBNET_INFO_VQ `ndr:"unique"`
	Status       ndr.DWORD                    `ndr:"retval"`
}

// R_DhcpGetSubnetInfoVQ calls R_DhcpGetSubnetInfoVQ (opnum 49) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetSubnetInfoVQ(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD) (SubnetInfoVQ *msdhcpm.DHCP_SUBNET_INFO_VQ, err error) {
	req := &r_DhcpGetSubnetInfoVQRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
	}
	var resp r_DhcpGetSubnetInfoVQResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetSubnetInfoVQ: %w", err)
		return
	}
	SubnetInfoVQ = resp.SubnetInfoVQ
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetSubnetInfoVQ failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
