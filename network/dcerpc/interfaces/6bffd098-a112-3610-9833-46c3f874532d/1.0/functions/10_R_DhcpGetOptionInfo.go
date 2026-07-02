package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetOptionInfoRequest carries the [in] parameters of R_DhcpGetOptionInfo.
type r_DhcpGetOptionInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	OptionID        ndr.DWORD
}

func (*r_DhcpGetOptionInfoRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpGetOptionInfo }

// r_DhcpGetOptionInfoResponse carries the [out] parameters and return value of R_DhcpGetOptionInfo.
type r_DhcpGetOptionInfoResponse struct {
	OptionInfo *msdhcpm.DHCP_OPTION `ndr:"unique"`
	Status     ndr.DWORD            `ndr:"retval"`
}

// R_DhcpGetOptionInfo calls R_DhcpGetOptionInfo (opnum 10) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetOptionInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, optionID ndr.DWORD) (OptionInfo *msdhcpm.DHCP_OPTION, err error) {
	req := &r_DhcpGetOptionInfoRequest{
		ServerIpAddress: serverIpAddress,
		OptionID:        optionID,
	}
	var resp r_DhcpGetOptionInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetOptionInfo: %w", err)
		return
	}
	OptionInfo = resp.OptionInfo
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetOptionInfo failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
