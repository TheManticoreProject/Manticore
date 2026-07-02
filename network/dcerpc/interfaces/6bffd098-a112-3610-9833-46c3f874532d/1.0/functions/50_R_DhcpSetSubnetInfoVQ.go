package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpSetSubnetInfoVQRequest carries the [in] parameters of R_DhcpSetSubnetInfoVQ.
type r_DhcpSetSubnetInfoVQRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
	SubnetInfoVQ    msdhcpm.DHCP_SUBNET_INFO_VQ
}

func (*r_DhcpSetSubnetInfoVQRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpSetSubnetInfoVQ }

// r_DhcpSetSubnetInfoVQResponse carries the [out] parameters and return value of R_DhcpSetSubnetInfoVQ.
type r_DhcpSetSubnetInfoVQResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetSubnetInfoVQ calls R_DhcpSetSubnetInfoVQ (opnum 50) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetSubnetInfoVQ(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, subnetInfoVQ msdhcpm.DHCP_SUBNET_INFO_VQ) (err error) {
	req := &r_DhcpSetSubnetInfoVQRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		SubnetInfoVQ:    subnetInfoVQ,
	}
	var resp r_DhcpSetSubnetInfoVQResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetSubnetInfoVQ: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetSubnetInfoVQ failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
