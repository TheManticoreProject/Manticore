package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpRemoveSubnetElementRequest carries the [in] parameters of R_DhcpRemoveSubnetElement.
type r_DhcpRemoveSubnetElementRequest struct {
	ServerIpAddress   *ndr.WSTR `ndr:"unique"`
	SubnetAddress     ndr.DWORD
	RemoveElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA
	ForceFlag         msdhcpm.DHCP_FORCE_FLAG
}

func (*r_DhcpRemoveSubnetElementRequest) Opnum() uint16 {
	return dhcpsrv.OpnumR_DhcpRemoveSubnetElement
}

// r_DhcpRemoveSubnetElementResponse carries the [out] parameters and return value of R_DhcpRemoveSubnetElement.
type r_DhcpRemoveSubnetElementResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpRemoveSubnetElement calls R_DhcpRemoveSubnetElement (opnum 6) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpRemoveSubnetElement(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, removeElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA, forceFlag msdhcpm.DHCP_FORCE_FLAG) (err error) {
	req := &r_DhcpRemoveSubnetElementRequest{
		ServerIpAddress:   serverIpAddress,
		SubnetAddress:     subnetAddress,
		RemoveElementInfo: removeElementInfo,
		ForceFlag:         forceFlag,
	}
	var resp r_DhcpRemoveSubnetElementResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpRemoveSubnetElement: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpRemoveSubnetElement failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
