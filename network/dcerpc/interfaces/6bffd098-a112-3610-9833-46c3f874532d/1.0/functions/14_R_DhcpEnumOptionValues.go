package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpEnumOptionValuesRequest carries the [in] parameters of R_DhcpEnumOptionValues.
type r_DhcpEnumOptionValuesRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	ScopeInfo        msdhcpm.DHCP_OPTION_SCOPE_INFO
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumOptionValuesRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpEnumOptionValues }

// r_DhcpEnumOptionValuesResponse carries the [out] parameters and return value of R_DhcpEnumOptionValues.
type r_DhcpEnumOptionValuesResponse struct {
	ResumeHandle ndr.DWORD
	OptionValues *msdhcpm.DHCP_OPTION_VALUE_ARRAY `ndr:"unique"`
	OptionsRead  ndr.DWORD
	OptionsTotal ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumOptionValues calls R_DhcpEnumOptionValues (opnum 14) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumOptionValues(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, OptionValues *msdhcpm.DHCP_OPTION_VALUE_ARRAY, OptionsRead ndr.DWORD, OptionsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumOptionValuesRequest{
		ServerIpAddress:  serverIpAddress,
		ScopeInfo:        scopeInfo,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumOptionValuesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumOptionValues: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	OptionValues = resp.OptionValues
	OptionsRead = resp.OptionsRead
	OptionsTotal = resp.OptionsTotal
	if uint32(resp.Status) != dhcpsrv.StatusSuccess && !dhcpsrv.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumOptionValues failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
