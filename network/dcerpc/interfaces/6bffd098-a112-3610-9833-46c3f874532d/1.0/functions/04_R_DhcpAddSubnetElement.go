package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpAddSubnetElementRequest carries the [in] parameters of R_DhcpAddSubnetElement.
type r_DhcpAddSubnetElementRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SubnetAddress   ndr.DWORD
	AddElementInfo  msdhcpm.DHCP_SUBNET_ELEMENT_DATA
}

func (*r_DhcpAddSubnetElementRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpAddSubnetElement }

// r_DhcpAddSubnetElementResponse carries the [out] parameters and return value of R_DhcpAddSubnetElement.
type r_DhcpAddSubnetElementResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpAddSubnetElement calls R_DhcpAddSubnetElement (opnum 4) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpAddSubnetElement(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, addElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA) (err error) {
	req := &r_DhcpAddSubnetElementRequest{
		ServerIpAddress: serverIpAddress,
		SubnetAddress:   subnetAddress,
		AddElementInfo:  addElementInfo,
	}
	var resp r_DhcpAddSubnetElementResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpAddSubnetElement: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpAddSubnetElement failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
