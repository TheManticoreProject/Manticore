package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpCreateOptionRequest carries the [in] parameters of R_DhcpCreateOption.
type r_DhcpCreateOptionRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	OptionID        ndr.DWORD
	OptionInfo      msdhcpm.DHCP_OPTION
}

func (*r_DhcpCreateOptionRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpCreateOption }

// r_DhcpCreateOptionResponse carries the [out] parameters and return value of R_DhcpCreateOption.
type r_DhcpCreateOptionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpCreateOption calls R_DhcpCreateOption (opnum 8) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpCreateOption(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, optionID ndr.DWORD, optionInfo msdhcpm.DHCP_OPTION) (err error) {
	req := &r_DhcpCreateOptionRequest{
		ServerIpAddress: serverIpAddress,
		OptionID:        optionID,
		OptionInfo:      optionInfo,
	}
	var resp r_DhcpCreateOptionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpCreateOption: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpCreateOption failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
