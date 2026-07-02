package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpSetOptionInfoRequest carries the [in] parameters of R_DhcpSetOptionInfo.
type r_DhcpSetOptionInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	OptionID        ndr.DWORD
	OptionInfo      msdhcpm.DHCP_OPTION
}

func (*r_DhcpSetOptionInfoRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpSetOptionInfo }

// r_DhcpSetOptionInfoResponse carries the [out] parameters and return value of R_DhcpSetOptionInfo.
type r_DhcpSetOptionInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetOptionInfo calls R_DhcpSetOptionInfo (opnum 9) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetOptionInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, optionID ndr.DWORD, optionInfo msdhcpm.DHCP_OPTION) (err error) {
	req := &r_DhcpSetOptionInfoRequest{
		ServerIpAddress: serverIpAddress,
		OptionID:        optionID,
		OptionInfo:      optionInfo,
	}
	var resp r_DhcpSetOptionInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetOptionInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetOptionInfo failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
