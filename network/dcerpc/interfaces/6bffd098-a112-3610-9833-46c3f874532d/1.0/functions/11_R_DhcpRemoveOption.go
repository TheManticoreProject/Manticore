package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpRemoveOptionRequest carries the [in] parameters of R_DhcpRemoveOption.
type r_DhcpRemoveOptionRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	OptionID        ndr.DWORD
}

func (*r_DhcpRemoveOptionRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpRemoveOption }

// r_DhcpRemoveOptionResponse carries the [out] parameters and return value of R_DhcpRemoveOption.
type r_DhcpRemoveOptionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpRemoveOption calls R_DhcpRemoveOption (opnum 11) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpRemoveOption(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, optionID ndr.DWORD) (err error) {
	req := &r_DhcpRemoveOptionRequest{
		ServerIpAddress: serverIpAddress,
		OptionID:        optionID,
	}
	var resp r_DhcpRemoveOptionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpRemoveOption: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpRemoveOption failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
