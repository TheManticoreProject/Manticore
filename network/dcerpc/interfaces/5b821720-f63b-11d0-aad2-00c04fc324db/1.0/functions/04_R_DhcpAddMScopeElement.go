package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpAddMScopeElementRequest carries the [in] parameters of R_DhcpAddMScopeElement.
type r_DhcpAddMScopeElementRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	MScopeName      ndr.WSTR
	AddElementInfo  msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V4
}

func (*r_DhcpAddMScopeElementRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpAddMScopeElement }

// r_DhcpAddMScopeElementResponse carries the [out] parameters and return value of R_DhcpAddMScopeElement.
type r_DhcpAddMScopeElementResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpAddMScopeElement calls R_DhcpAddMScopeElement (opnum 4) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpAddMScopeElement(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, mScopeName ndr.WSTR, addElementInfo msdhcpm.DHCP_SUBNET_ELEMENT_DATA_V4) (err error) {
	req := &r_DhcpAddMScopeElementRequest{
		ServerIpAddress: serverIpAddress,
		MScopeName:      mScopeName,
		AddElementInfo:  addElementInfo,
	}
	var resp r_DhcpAddMScopeElementResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpAddMScopeElement: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpAddMScopeElement failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
