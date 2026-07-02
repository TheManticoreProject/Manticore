package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpSetServerBindingInfoRequest carries the [in] parameters of R_DhcpSetServerBindingInfo.
type r_DhcpSetServerBindingInfoRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	Flags            ndr.DWORD
	BindElementsInfo msdhcpm.DHCP_BIND_ELEMENT_ARRAY
}

func (*r_DhcpSetServerBindingInfoRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpSetServerBindingInfo
}

// r_DhcpSetServerBindingInfoResponse carries the [out] parameters and return value of R_DhcpSetServerBindingInfo.
type r_DhcpSetServerBindingInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetServerBindingInfo calls R_DhcpSetServerBindingInfo (opnum 41) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetServerBindingInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, bindElementsInfo msdhcpm.DHCP_BIND_ELEMENT_ARRAY) (err error) {
	req := &r_DhcpSetServerBindingInfoRequest{
		ServerIpAddress:  serverIpAddress,
		Flags:            flags,
		BindElementsInfo: bindElementsInfo,
	}
	var resp r_DhcpSetServerBindingInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetServerBindingInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetServerBindingInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
