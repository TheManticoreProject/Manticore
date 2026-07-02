package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpEnumOptionsRequest carries the [in] parameters of R_DhcpEnumOptions.
type r_DhcpEnumOptionsRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumOptionsRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpEnumOptions }

// r_DhcpEnumOptionsResponse carries the [out] parameters and return value of R_DhcpEnumOptions.
type r_DhcpEnumOptionsResponse struct {
	ResumeHandle ndr.DWORD
	Options      *msdhcpm.DHCP_OPTION_ARRAY `ndr:"unique"`
	OptionsRead  ndr.DWORD
	OptionsTotal ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumOptions calls R_DhcpEnumOptions (opnum 23) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumOptions(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, Options *msdhcpm.DHCP_OPTION_ARRAY, OptionsRead ndr.DWORD, OptionsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumOptionsRequest{
		ServerIpAddress:  serverIpAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumOptionsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumOptions: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	Options = resp.Options
	OptionsRead = resp.OptionsRead
	OptionsTotal = resp.OptionsTotal
	if uint32(resp.Status) != dhcpsrv.StatusSuccess && !dhcpsrv.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumOptions failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
