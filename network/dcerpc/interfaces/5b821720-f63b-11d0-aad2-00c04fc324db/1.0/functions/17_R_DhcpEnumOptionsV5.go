package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpEnumOptionsV5Request carries the [in] parameters of R_DhcpEnumOptionsV5.
type r_DhcpEnumOptionsV5Request struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	Flags            ndr.DWORD
	ClassName        *ndr.WSTR `ndr:"unique"`
	VendorName       *ndr.WSTR `ndr:"unique"`
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumOptionsV5Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpEnumOptionsV5 }

// r_DhcpEnumOptionsV5Response carries the [out] parameters and return value of R_DhcpEnumOptionsV5.
type r_DhcpEnumOptionsV5Response struct {
	ResumeHandle ndr.DWORD
	Options      *msdhcpm.DHCP_OPTION_ARRAY `ndr:"unique"`
	OptionsRead  ndr.DWORD
	OptionsTotal ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumOptionsV5 calls R_DhcpEnumOptionsV5 (opnum 17) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumOptionsV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, Options *msdhcpm.DHCP_OPTION_ARRAY, OptionsRead ndr.DWORD, OptionsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumOptionsV5Request{
		ServerIpAddress:  serverIpAddress,
		Flags:            flags,
		ClassName:        className,
		VendorName:       vendorName,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumOptionsV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumOptionsV5: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	Options = resp.Options
	OptionsRead = resp.OptionsRead
	OptionsTotal = resp.OptionsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumOptionsV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
