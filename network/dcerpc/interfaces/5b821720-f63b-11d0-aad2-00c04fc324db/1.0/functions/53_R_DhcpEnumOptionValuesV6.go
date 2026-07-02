package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpEnumOptionValuesV6Request carries the [in] parameters of R_DhcpEnumOptionValuesV6.
type r_DhcpEnumOptionValuesV6Request struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	Flags            ndr.DWORD
	ClassName        *ndr.WSTR `ndr:"unique"`
	VendorName       *ndr.WSTR `ndr:"unique"`
	ScopeInfo        msdhcpm.DHCP_OPTION_SCOPE_INFO6
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumOptionValuesV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpEnumOptionValuesV6 }

// r_DhcpEnumOptionValuesV6Response carries the [out] parameters and return value of R_DhcpEnumOptionValuesV6.
type r_DhcpEnumOptionValuesV6Response struct {
	ResumeHandle ndr.DWORD
	OptionValues *msdhcpm.DHCP_OPTION_VALUE_ARRAY `ndr:"unique"`
	OptionsRead  ndr.DWORD
	OptionsTotal ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumOptionValuesV6 calls R_DhcpEnumOptionValuesV6 (opnum 53) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumOptionValuesV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO6, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, OptionValues *msdhcpm.DHCP_OPTION_VALUE_ARRAY, OptionsRead ndr.DWORD, OptionsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumOptionValuesV6Request{
		ServerIpAddress:  serverIpAddress,
		Flags:            flags,
		ClassName:        className,
		VendorName:       vendorName,
		ScopeInfo:        scopeInfo,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumOptionValuesV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumOptionValuesV6: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	OptionValues = resp.OptionValues
	OptionsRead = resp.OptionsRead
	OptionsTotal = resp.OptionsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumOptionValuesV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
