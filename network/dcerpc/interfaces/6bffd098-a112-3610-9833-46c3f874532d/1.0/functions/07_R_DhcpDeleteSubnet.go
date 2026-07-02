package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpDeleteSubnetRequest carries the [in] parameters of R_DhcpDeleteSubnet.
type r_DhcpDeleteSubnetRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
	ForceFlag       msdhcpm.DHCP_FORCE_FLAG
}

func (*r_DhcpDeleteSubnetRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpDeleteSubnet }

// r_DhcpDeleteSubnetResponse carries the [out] parameters and return value of R_DhcpDeleteSubnet.
type r_DhcpDeleteSubnetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpDeleteSubnet calls R_DhcpDeleteSubnet (opnum 7) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpDeleteSubnet(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, forceFlag msdhcpm.DHCP_FORCE_FLAG) (err error) {
	req := &r_DhcpDeleteSubnetRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		ForceFlag:       forceFlag,
	}
	var resp r_DhcpDeleteSubnetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpDeleteSubnet: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpDeleteSubnet failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
